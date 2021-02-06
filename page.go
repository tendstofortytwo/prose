package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type page struct {
	slug string
}

func (p *page) render() ([]byte, error) {
	data, err := ioutil.ReadFile("posts/" + p.slug + ".md")
	if err != nil {
		return nil, fmt.Errorf("Could not read from %s.md: %s", p.slug, err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.Linkify),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var converted bytes.Buffer
	err = md.Convert(data, &converted)
	if err != nil {
		return nil, fmt.Errorf("Could not parse markdown from %s.md: %s", p.slug, err)
	}

	return converted.Bytes(), nil
}
