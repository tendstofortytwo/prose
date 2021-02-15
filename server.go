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
}

func newServer() (*server, error) {
	files, err := ioutil.ReadDir("posts/")

	if err != nil {
		return nil, err
	}

	s := &server{
		pages:         make([]*Page, 0, len(files)),
		staticHandler: http.FileServer(http.Dir("static/")),
	}

	for _, f := range files {
		filename := f.Name()

		if strings.HasSuffix(filename, ".md") {
			page, err := newPage(strings.TrimSuffix(filename, ".md"))
			if err != nil {
				return nil, fmt.Errorf("could not render %s: %s", filename, err)
			}
			s.pages = append(s.pages, page)
			log.Printf("Loaded page %s", filename)
		}
	}

	s.pageTpl, err = raymond.ParseFile("templates/page.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse page template")
	}
	log.Printf("Loaded page template")

	s.fullPostTpl, err = raymond.ParseFile("templates/fullpost.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse full post template")
	}
	log.Printf("Loaded full post template")

	s.summaryTpl, err = raymond.ParseFile("templates/summary.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse summary template")
	}
	log.Printf("Loaded summary template")

	styles, err := ioutil.ReadDir("styles/")
	if err != nil {
		return nil, fmt.Errorf("Could not load styles directory: %s", err)
	}

	for _, s := range styles {
		filename := s.Name()
		in, err := os.Open("styles/" + filename)
		if err != nil {
			return nil, fmt.Errorf("Could not open style infile %s: %s", filename, err)
		}
		if strings.HasSuffix(filename, ".scss") {
			outFilename := strings.TrimSuffix(filename, ".scss") + ".css"
			out, err := os.Create("static/css/" + outFilename)
			if err != nil {
				return nil, fmt.Errorf("Could not open style outfile %s: %s", outFilename, err)
			}
			comp, err := libsass.New(out, in)
			if err != nil {
				return nil, fmt.Errorf("Could not start sass compiler for file %s: %s", filename, err)
			}
			if err = comp.Run(); err != nil {
				return nil, fmt.Errorf("Could not generate stylesheet %s: %s", filename, err)
			}
		} else if strings.HasSuffix(filename, ".css") {
			out, err := os.Create("static/css/" + filename)
			if err != nil {
				return nil, fmt.Errorf("Could not open style outfile %s: %s", filename, err)
			}
			_, err = io.Copy(out, in)
			if err != nil {
				return nil, fmt.Errorf("Could not copy stylesheet %s: %s", filename, err)
			}
		} else {
			log.Printf("Skipping stylesheet %s, don't know how to handle", filename)
			continue
		}
		log.Printf("Loaded stylesheet %s", filename)
	}

	return s, nil
}

func (s *server) logRequest(req *http.Request) {
	log.Printf("%s %s from %s", req.Method, req.URL.Path, req.RemoteAddr)
}

func (s *server) router(res http.ResponseWriter, req *http.Request) {
	s.logRequest(req)
	slug := req.URL.Path[1:]

	if slug == "" {
		s.homePage(res, req)
		return
	}

	for _, p := range s.pages {
		if p.Slug == slug {
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
