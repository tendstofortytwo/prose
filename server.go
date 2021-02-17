package main

import (
	"fmt"
	"io"
	"io/ioutil"
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
	files, err := ioutil.ReadDir("posts/")
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

func (s *server) refreshTemplates() error {
	var err error
	s.pageTpl, err = raymond.ParseFile("templates/page.html")
	if err != nil {
		return fmt.Errorf("could not parse page template")
	}
	log.Printf("Loaded page template")

	s.fullPostTpl, err = raymond.ParseFile("templates/fullpost.html")
	if err != nil {
		return fmt.Errorf("could not parse full post template")
	}
	log.Printf("Loaded full post template")

	s.notFoundTpl, err = raymond.ParseFile("templates/notfound.html")
	if err != nil {
		return fmt.Errorf("could not parse 404 template")
	}
	log.Printf("Loaded 404 template")

	s.errorTpl, err = raymond.ParseFile("templates/error.html")
	if err != nil {
		return fmt.Errorf("could not parse error template")
	}
	log.Printf("Loaded error template")

	s.summaryTpl, err = raymond.ParseFile("templates/summary.html")
	if err != nil {
		return fmt.Errorf("could not parse summary template")
	}
	log.Printf("Loaded summary template")
	return nil
}

func (s *server) refreshStyles() error {
	styles, err := ioutil.ReadDir("styles/")
	if err != nil {
		return fmt.Errorf("Could not load styles directory: %s", err)
	}

	for _, s := range styles {
		filename := s.Name()
		in, err := os.Open("styles/" + filename)
		if err != nil {
			return fmt.Errorf("Could not open style infile %s: %s", filename, err)
		}
		if strings.HasSuffix(filename, ".scss") {
			outFilename := strings.TrimSuffix(filename, ".scss") + ".css"
			out, err := os.Create("static/css/" + outFilename)
			if err != nil {
				return fmt.Errorf("Could not open style outfile %s: %s", outFilename, err)
			}
			comp, err := libsass.New(out, in)
			if err != nil {
				return fmt.Errorf("Could not start sass compiler for file %s: %s", filename, err)
			}
			if err = comp.Run(); err != nil {
				return fmt.Errorf("Could not generate stylesheet %s: %s", filename, err)
			}
		} else if strings.HasSuffix(filename, ".css") {
			out, err := os.Create("static/css/" + filename)
			if err != nil {
				return fmt.Errorf("Could not open style outfile %s: %s", filename, err)
			}
			_, err = io.Copy(out, in)
			if err != nil {
				return fmt.Errorf("Could not copy stylesheet %s: %s", filename, err)
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
