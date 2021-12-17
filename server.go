package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/aymerick/raymond"
	"github.com/fogleman/gg"
)

const (
	blogTitle   = "Prose"
	blogURL     = "https://prose.nsood.in"
	blogSummary = "Where I infodump in Markdown and nobody can stop me."
)

type server struct {
	staticHandler http.Handler

	mu        sync.RWMutex
	templates map[string]*raymond.Template
	postList
	styles    map[string]string
	homeImage []byte
}

func newServer() (*server, error) {
	s := &server{
		staticHandler: http.FileServer(http.Dir("static/")),
	}

	var imgBuffer bytes.Buffer
	err := createImage(blogTitle, blogSummary, blogURL, &imgBuffer)
	if err != nil {
		return nil, err
	}
	s.homeImage, err = io.ReadAll(&imgBuffer)
	if err != nil {
		return nil, err
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
	if slug == "about.png" {
		s.renderImage(res, req, s.homeImage)
		return
	}

	for _, p := range s.postList {
		if slug == p.Slug {
			s.postPage(p, res, req)
			return
		} else if slug == p.Slug+"/about.png" {
			s.renderImage(res, req, p.Image)
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

func (s *server) createWebPage(title, subtitle, contents, path string) (string, error) {
	ctx := map[string]interface{}{
		"title":    title,
		"subtitle": subtitle,
		"contents": contents,
		"path":     blogURL + path,
	}
	return s.templates["page"].Exec(ctx)
}

func (s *server) postPage(p *Post, res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html; charset=utf-8")
	contents, err := p.render(s.templates["fullpost"])
	if err != nil {
		s.errorInRequest(res, req, err)
	}
	page, err := s.createWebPage(p.Metadata.Title, p.Metadata.Summary, contents, req.URL.Path)
	if err != nil {
		s.errorInRequest(res, req, err)
	}
	res.Write([]byte(page))
}

func (s *server) homePage(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("content-type", "text/html; charset=utf-8")

	var posts string

	for _, p := range s.postList {
		summary, err := p.render(s.templates["summary"])
		if err != nil {
			log.Printf("could not render post summary for %s", p.Slug)
		}
		posts = posts + summary
	}

	page, err := s.createWebPage("Home", blogSummary, posts, "")

	if err != nil {
		s.errorInRequest(res, req, err)
	}

	res.Write([]byte(page))
}

func (s *server) renderImage(res http.ResponseWriter, req *http.Request, img []byte) {
	res.Header().Add("content-type", "image/png")
	res.Write(img)
}

func (s *server) loadStylesheet(res http.ResponseWriter, req *http.Request, filename string) (ok bool) {
	contents, ok := s.styles[filename]
	if !ok {
		return false
	}
	res.Header().Add("content-type", "text/css")
	res.Write([]byte(contents))
	return ok
}

func createImage(title, summary, url string, out io.Writer) error {
	imgWidth, imgHeight, imgPaddingX, imgPaddingY := 1200, 600, 50, 100
	accentHeight, spacerHeight := 12.5, 20.0
	titleSize, summarySize, urlSize := 63.0, 42.0, 27.0
	lineHeight := 1.05
	textWidth := float64(imgWidth - 2*imgPaddingX)

	draw := gg.NewContext(imgWidth, imgHeight)

	titleFont, err := gg.LoadFontFace("fonts/Nunito-Bold.ttf", titleSize)
	if err != nil {
		return err
	}
	summaryFont, err := gg.LoadFontFace("fonts/Nunito-LightItalic.ttf", summarySize)
	if err != nil {
		return err
	}
	urlFont, err := gg.LoadFontFace("fonts/JetBrainsMono-ExtraLight.ttf", urlSize)
	if err != nil {
		return err
	}

	draw.SetFontFace(titleFont)
	wrappedTitle := draw.WordWrap(title, textWidth)
	draw.SetFontFace(summaryFont)
	wrappedSummary := draw.WordWrap(summary, textWidth)

	draw.SetHexColor("#fff")
	draw.DrawRectangle(0, 0, float64(imgWidth), float64(imgHeight))
	draw.Fill()
	draw.SetHexColor("#3498db")
	draw.DrawRectangle(0, float64(imgHeight)-accentHeight, float64(imgWidth), accentHeight)
	draw.Fill()

	offset := float64(imgPaddingY)

	draw.SetFontFace(titleFont)
	draw.SetHexColor("#333")
	for _, line := range wrappedTitle {
		draw.DrawString(line, float64(imgPaddingX), offset)
		offset += lineHeight * titleSize
	}
	offset += spacerHeight

	draw.SetFontFace(summaryFont)
	draw.SetHexColor("#999")
	for _, line := range wrappedSummary {
		draw.DrawString(line, float64(imgPaddingX), offset)
		offset += lineHeight * summarySize
	}

	draw.SetHexColor("#333")
	draw.SetFontFace(urlFont)
	urlY := float64(imgHeight - imgPaddingY)
	draw.DrawStringWrapped(url, float64(imgPaddingX), urlY, 0, 0, textWidth, lineHeight, gg.AlignRight)

	return draw.EncodePNG(out)
}
