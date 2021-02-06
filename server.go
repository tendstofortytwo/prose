package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type server struct {
	pages         []page
	staticHandler http.Handler
}

func newServer() (*server, error) {
	files, err := ioutil.ReadDir("posts/")

	if err != nil {
		return nil, err
	}

	s := &server{
		pages:         make([]page, 0, len(files)),
		staticHandler: http.FileServer(http.Dir("static/")),
	}

	for _, f := range files {
		filename := f.Name()

		if strings.HasSuffix(filename, ".md") {
			s.pages = append(s.pages, page{slug: strings.TrimSuffix(filename, ".md")})
		}
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
		if p.slug == slug {
			buf, err := p.render()

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				res.Write([]byte("oh no"))
			}

			res.Write(buf)

			return
		}
	}

	s.staticHandler.ServeHTTP(res, req)
}

func (s *server) homePage(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html")

	res.Write([]byte("<h1>blog</h1>"))

	for _, p := range s.pages {
		fmt.Fprintf(res, "<a href=\"/%s\">%s</a><br>", p.slug, p.slug)
	}
}
