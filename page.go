package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aymerick/raymond"
	"github.com/mitchellh/mapstructure"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// Metadata stores the data about a page that needs to be visible
// at the home page.
type Metadata struct {
	Title   string
	Summary string
	Date    string // TODO: better representation? time.Time might cause timezone issues...
}

// Page stores the contents of a blog post.
type Page struct {
	Slug     string
	Metadata Metadata
	Contents string
}

func newPage(slug string) (*Page, error) {
	data, err := os.ReadFile("posts/" + slug + ".md")
	if err != nil {
		return nil, fmt.Errorf("could not read file: %s", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Strikethrough,
			extension.Typographer,
			meta.Meta,
			highlighting.Highlighting,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var converted bytes.Buffer
	ctx := parser.NewContext()
	err = md.Convert(data, &converted, parser.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("could not parse markdown: %s", err)
	}
	mdMap, err := meta.TryGet(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not parse metadata: %s", err)
	}
	var metadata Metadata
	err = mapstructure.Decode(mdMap, &metadata)
	if err != nil {
		return nil, fmt.Errorf("could not destructure metadata: %s", err)
	}

	page := &Page{
		Slug:     slug,
		Metadata: metadata,
		Contents: converted.String(),
	}

	return page, nil
}

func (p *Page) render(tpl *raymond.Template) (string, error) {
	return tpl.Exec(p)
}
