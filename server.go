package main

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
)

type server struct {
	staticHandler http.Handler

	mu        sync.RWMutex
	templates map[string]*raymond.Template
	postList
	styles map[string]string
}

func newServer() (*server, error) {
	s := &server{
		staticHandler: http.FileServer(http.Dir("static/")),
	}

	posts, err := newPostList()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.postList = posts
	s.mu.Unlock()

	tpls, err := loadTemplates([]string{"page", "fullpost", "summary", "notfound", "error"})
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.templates = tpls
	s.mu.Unlock()

	styles, err := newStylesMap()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.styles = styles
	s.mu.Unlock()

	postsLn := newPostListener(func(updateFn func(postList) postList) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.postList = updateFn(s.postList)
	})
	go postsLn.listen()

	templatesLn := newTemplateListener(func(updateFn func(map[string]*raymond.Template)) {
		s.mu.Lock()
		defer s.mu.Unlock()
		updateFn(s.templates)
	})
	go templatesLn.listen()

	stylesLn := newStylesListener(func(updateFn func(map[string]string)) {
		s.mu.Lock()
		defer s.mu.Unlock()
		updateFn(s.styles)
	})
	go stylesLn.listen()

	return s, nil
}

func (s *server) logRequest(req *http.Request) {
	log.Printf("%s %s from %s", req.Method, req.URL.Path, req.RemoteAddr)
}

func (s *server) router(res http.ResponseWriter, req *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

	s.mu.RLock()
	defer s.mu.RUnlock()
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

	s.mu.RLock()
	defer s.mu.RUnlock()
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	contents, ok := s.styles[filename]
	if !ok {
		return false
	}
	res.Header().Add("content-type", "text/css")
	res.Write([]byte(contents))
	return ok
}
