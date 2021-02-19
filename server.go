package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aymerick/raymond"
	"github.com/wellington/go-libsass"
)

type server struct {
	pages         []*Page
	staticHandler http.Handler
	pageTpl       *raymond.Template
	fullPostTpl   *raymond.Template
	summaryTpl    *raymond.Template
	errorTpl      *raymond.Template
	notFoundTpl   *raymond.Template
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

// loadTemplates, for each f in files, loads `templates/$f.html`
// as a handlebars HTML template. If any single template fails to
// load, only an error is returned. Conversely, if there is no error,
// every template name passed is guaranteed to have loaded successfully.
func loadTemplates(files []string) ([]*raymond.Template, error) {
	templates := make([]*raymond.Template, 0, len(files))
	for _, f := range files {
		tpl, err := raymond.ParseFile("templates/" + f + ".html")
		if err != nil {
			return nil, fmt.Errorf("Could not parse %s template: %w", f, err)
		}
		templates = append(templates, tpl)
	}
	log.Printf("Loaded templates: %s", files)
	return templates, nil
}

func (s *server) refreshTemplates() error {
	templates, err := loadTemplates([]string{"page", "fullpost", "summary", "notfound", "error"})
	if err != nil {
		return err
	}
	s.pageTpl = templates[0]
	s.fullPostTpl = templates[1]
	s.summaryTpl = templates[2]
	s.notFoundTpl = templates[3]
	s.errorTpl = templates[4]
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

func (s *server) logRequest(req *http.Request) {
	log.Printf("%s %s from %s", req.Method, req.URL.Path, req.RemoteAddr)
}

func (s *server) router(res http.ResponseWriter, req *http.Request) {
	s.logRequest(req)
	res = &errorCatcher{
		res:          res,
		req:          req,
		errorTpl:     s.errorTpl,
		notFoundTpl:  s.notFoundTpl,
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
	return s.pageTpl.Exec(ctx)
}

func (s *server) renderPage(p *Page, res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")
	contents, err := p.render(s.fullPostTpl)
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
		summary, err := p.render(s.summaryTpl)
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
