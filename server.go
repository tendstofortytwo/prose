package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
	"github.com/fsnotify/fsnotify"
	"github.com/wellington/go-libsass"
)

type server struct {
	pages         []*Page
	staticHandler http.Handler

	tplMutex  sync.RWMutex
	templates map[string]*raymond.Template
}

func newServer() (*server, error) {
	s := &server{
		staticHandler: http.FileServer(http.Dir("static/")),
	}

	err := s.refreshPages()
	if err != nil {
		return nil, err
	}

	err = s.refreshTemplates()
	if err != nil {
		return nil, err
	}

	err = s.refreshStyles()
	if err != nil {
		return nil, err
	}

	go listenForChanges("templates/", func(file string) error {
		s.tplMutex.Lock()
		defer s.tplMutex.Unlock()
		tplName := strings.TrimSuffix(file, ".html")
		newTpl, err := loadTemplate(tplName)
		if err != nil {
			return err
		}
		s.templates[tplName] = newTpl
		return nil
	})

	go listenForChanges("styles/", func(file string) error {
		var err error
		if strings.HasSuffix(file, ".scss") {
			err = loadSassStylesheet(file)
		} else if strings.HasSuffix(file, ".css") {
			err = loadRegularStylesheet(file)
		}
		return err
	})

	return s, nil
}

func (s *server) refreshPages() error {
	files, err := os.ReadDir("posts/")
	if err != nil {
		return err
	}

	s.pages = make([]*Page, 0, len(files))
	for _, f := range files {
		filename := f.Name()

		if strings.HasSuffix(filename, ".md") {
			page, err := newPage(strings.TrimSuffix(filename, ".md"))
			if err != nil {
				return fmt.Errorf("could not render %s: %s", filename, err)
			}
			s.pages = append(s.pages, page)
			log.Printf("Loaded page %s", filename)
		}
	}

	return nil
}

func loadTemplate(file string) (*raymond.Template, error) {
	tpl, err := raymond.ParseFile("templates/" + file + ".html")
	if err != nil {
		return nil, fmt.Errorf("Could not parse %s template: %w", file, err)
	}
	log.Printf("Loaded template: %s", file)
	return tpl, nil
}

// loadTemplates, for each f in files, loads `templates/$f.html`
// as a handlebars HTML template. If any single template fails to
// load, only an error is returned. Conversely, if there is no error,
// every template name passed is guaranteed to have loaded successfully.
func loadTemplates(files []string) (map[string]*raymond.Template, error) {
	templates := make(map[string]*raymond.Template)
	for _, f := range files {
		tpl, err := loadTemplate(f)
		if err != nil {
			return nil, err
		}
		templates[f] = tpl
	}
	log.Printf("Loaded templates: %s", files)
	return templates, nil
}

func (s *server) refreshTemplates() error {
	s.tplMutex.Lock()
	defer s.tplMutex.Unlock()
	tpls, err := loadTemplates([]string{"page", "fullpost", "summary", "notfound", "error"})
	if err != nil {
		return err
	}
	s.templates = tpls
	return nil
}

func loadSassStylesheet(filename string) error {
	in, err := os.Open("styles/" + filename)
	if err != nil {
		return fmt.Errorf("Could not open style infile %s: %w", filename, err)
	}
	output := strings.TrimSuffix(filename, ".scss") + ".css"
	out, err := os.Create("static/css/" + output)
	if err != nil {
		return fmt.Errorf("Could not open style outfile %s: %w", output, err)
	}
	comp, err := libsass.New(out, in)
	if err != nil {
		return fmt.Errorf("Could not start sass compiler for file %s: %w", filename, err)
	}
	if err = comp.Run(); err != nil {
		return fmt.Errorf("Could not generate stylesheet %s: %w", filename, err)
	}
	return nil
}

func loadRegularStylesheet(filename string) error {
	in, err := os.Open("styles/" + filename)
	if err != nil {
		return fmt.Errorf("Could not open style infile %s: %w", filename, err)
	}
	out, err := os.Create("static/css/" + filename)
	if err != nil {
		return fmt.Errorf("Could not open style outfile %s: %w", filename, err)
	}
	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("Could not copy stylesheet %s: %s", filename, err)
	}
	return nil
}

func (s *server) refreshStyles() error {
	styles, err := os.ReadDir("styles/")
	if err != nil {
		return fmt.Errorf("Could not load styles directory: %s", err)
	}

	for _, s := range styles {
		filename := s.Name()
		if strings.HasSuffix(filename, ".scss") {
			err := loadSassStylesheet(filename)
			if err != nil {
				return err
			}
		} else if strings.HasSuffix(filename, ".css") {
			err := loadRegularStylesheet(filename)
			if err != nil {
				return err
			}
		} else {
			log.Printf("Skipping stylesheet %s, don't know how to handle", filename)
			continue
		}
		log.Printf("Loaded stylesheet %s", filename)
	}
	return nil
}

func listenForChanges(folder string, action func(string) error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("could not start fsnotify watcher for folder %s", folder)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Modified file: %s", event.Name)
					err := action(strings.TrimPrefix(event.Name, folder))
					if err != nil {
						log.Printf("watcher action on %s failed: %v", event.Name, err)
					}
				}
			case err := <-watcher.Errors:
				log.Printf("Watcher error: %s", err)
			}
		}
	}()

	err = watcher.Add(folder)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (s *server) logRequest(req *http.Request) {
	log.Printf("%s %s from %s", req.Method, req.URL.Path, req.RemoteAddr)
}

func (s *server) router(res http.ResponseWriter, req *http.Request) {
	s.tplMutex.RLock()
	defer s.tplMutex.RUnlock()
	s.logRequest(req)
	res = &errorCatcher{
		res:          res,
		req:          req,
		errorTpl:     s.templates["error"],
		notFoundTpl:  s.templates["notfound"],
		handledError: false,
	}
	slug := req.URL.Path[1:]

	if slug == "" {
		s.homePage(res, req)
		return
	}

	for _, p := range s.pages {
		if p.Slug == slug {
			s.renderPage(p, res, req)
			return
		}
	}

	s.staticHandler.ServeHTTP(res, req)
}

func (s *server) errorInRequest(res http.ResponseWriter, req *http.Request, err error) {
	res.WriteHeader(http.StatusInternalServerError)
	res.Write([]byte("oh no"))
	log.Printf("ERR %s: %s", req.URL.Path, err)
}

func (s *server) createPage(title, contents string) (string, error) {
	ctx := map[string]interface{}{
		"title":    title,
		"contents": contents,
	}
	return s.templates["page"].Exec(ctx)
}

func (s *server) renderPage(p *Page, res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")
	contents, err := p.render(s.templates["fullpost"])
	if err != nil {
		s.errorInRequest(res, req, err)
	}
	page, err := s.createPage(p.Metadata.Title, contents)
	if err != nil {
		s.errorInRequest(res, req, err)
	}
	res.Write([]byte(page))
}

func (s *server) homePage(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")

	var posts string

	for _, p := range s.pages {
		summary, err := p.render(s.templates["summary"])
		if err != nil {
			log.Printf("could not render page summary for %s", p.Slug)
		}
		posts = posts + summary
	}

	page, err := s.createPage("Home", posts)

	if err != nil {
		s.errorInRequest(res, req, err)
	}

	res.Write([]byte(page))
}
