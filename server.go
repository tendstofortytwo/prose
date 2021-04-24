package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
	"github.com/rjeczalik/notify"
)

type server struct {
	staticHandler http.Handler

	tplMutex  sync.RWMutex
	templates map[string]*raymond.Template

	postsMutex sync.RWMutex
	postList

	cssMutex sync.RWMutex
	styles   map[string]string
}

func newServer() (*server, error) {
	s := &server{
		staticHandler: http.FileServer(http.Dir("static/")),
	}

	posts, err := newPostList()
	if err != nil {
		return nil, err
	}
	s.postsMutex.Lock()
	s.postList = posts
	s.postsMutex.Unlock()

	tpls, err := loadTemplates([]string{"page", "fullpost", "summary", "notfound", "error"})
	if err != nil {
		return nil, err
	}
	s.tplMutex.Lock()
	s.templates = tpls
	s.tplMutex.Unlock()

	styles, err := newStylesMap()
	if err != nil {
		return nil, err
	}
	s.cssMutex.Lock()
	s.styles = styles
	s.cssMutex.Unlock()

	postsLn := newPostListener(func(updateFn func(postList) postList) {
		s.postsMutex.Lock()
		defer s.postsMutex.Unlock()
		s.postList = updateFn(s.postList)
	})
	go postsLn.listen()

	templatesLn := newTemplateListener(func(updateFn func(map[string]*raymond.Template)) {
		s.tplMutex.Lock()
		defer s.tplMutex.Unlock()
		updateFn(s.templates)
	})
	go templatesLn.listen()

	stylesLn := newStylesListener(func(updateFn func(map[string]string)) {
		s.cssMutex.Lock()
		defer s.cssMutex.Unlock()
		updateFn(s.styles)
	})
	go stylesLn.listen()

	return s, nil
}

type listener struct {
	folder string
	update func(string) error
	clean  func(string) error
}

func (l *listener) listen() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("could not get current working directory for listener!")
	}
	cwd = cwd + "/"

	c := make(chan notify.EventInfo, 1)

	var events []notify.Event

	// inotify events prevent double-firing of
	// certain events in Linux.
	if runtime.GOOS == "linux" {
		events = []notify.Event{
			notify.InCloseWrite,
			notify.InMovedFrom,
			notify.InMovedTo,
			notify.InDelete,
		}
	} else {
		events = []notify.Event{
			notify.Create,
			notify.Remove,
			notify.Rename,
			notify.Write,
		}
	}

	err = notify.Watch(l.folder, c, events...)

	if err != nil {
		log.Fatalf("Could not setup watcher for folder %s: %s", l.folder, err)
	}

	defer notify.Stop(c)

	for {
		ei := <-c
		log.Printf("event: %s", ei.Event())
		switch ei.Event() {
		case notify.InCloseWrite, notify.InMovedTo, notify.Create, notify.Rename, notify.Write:
			filePath := strings.TrimPrefix(ei.Path(), cwd)
			log.Printf("updating file %s", filePath)
			err := l.update(strings.TrimPrefix(filePath, l.folder))
			if err != nil {
				log.Printf("watcher update action on %s failed: %v", filePath, err)
			}
		case notify.InMovedFrom, notify.InDelete, notify.Remove:
			filePath := strings.TrimPrefix(ei.Path(), cwd)
			log.Printf("cleaning file %s", filePath)
			err := l.clean(strings.TrimPrefix(filePath, l.folder))
			if err != nil {
				log.Printf("watcher clean action on %s failed: %v", filePath, err)
			}
		}
	}
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

	s.postsMutex.RLock()
	defer s.postsMutex.RUnlock()
	for _, p := range s.postList {
		if p.Slug == slug {
			s.postPage(p, res, req)
			return
		}
	}

	if strings.HasPrefix(slug, "css/") {
		filename := strings.TrimPrefix(slug, "css/")
		ok := s.loadStylesheet(res, req, filename)
		if ok {
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

func (s *server) createWebPage(title, contents string) (string, error) {
	ctx := map[string]interface{}{
		"title":    title,
		"contents": contents,
	}
	return s.templates["page"].Exec(ctx)
}

func (s *server) postPage(p *Post, res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")
	contents, err := p.render(s.templates["fullpost"])
	if err != nil {
		s.errorInRequest(res, req, err)
	}
	page, err := s.createWebPage(p.Metadata.Title, contents)
	if err != nil {
		s.errorInRequest(res, req, err)
	}
	res.Write([]byte(page))
}

func (s *server) homePage(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")

	var posts string

	s.postsMutex.RLock()
	defer s.postsMutex.RUnlock()
	for _, p := range s.postList {
		summary, err := p.render(s.templates["summary"])
		if err != nil {
			log.Printf("could not render post summary for %s", p.Slug)
		}
		posts = posts + summary
	}

	page, err := s.createWebPage("Home", posts)

	if err != nil {
		s.errorInRequest(res, req, err)
	}

	res.Write([]byte(page))
}

func (s *server) loadStylesheet(res http.ResponseWriter, req *http.Request, filename string) (ok bool) {
	s.cssMutex.RLock()
	defer s.cssMutex.RUnlock()
	contents, ok := s.styles[filename]
	if !ok {
		return false
	}
	res.Header().Add("content-type", "text/css")
	res.Write([]byte(contents))
	return ok
}
