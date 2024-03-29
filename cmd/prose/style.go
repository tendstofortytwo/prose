package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bep/godartsass/v2"
)

var sassTranspiler *godartsass.Transpiler

func newStylesMap() (map[string]string, error) {
	folder, err := os.ReadDir("styles/")
	if err != nil {
		return nil, fmt.Errorf("could not load styles directory: %s", err)
	}

	styles := make(map[string]string)
	for _, s := range folder {
		contents, filename, err := loadStylesheet(s.Name())
		if err != nil {
			return nil, fmt.Errorf("could not generate styles for %s: %v", s.Name(), err)
		}
		styles[filename] = contents
		log.Printf("Loaded stylesheet %s", filename)
	}

	return styles, nil
}

func newStylesListener(updateMap func(func(map[string]string))) *listener {
	ln := &listener{
		folder: "styles/",
		update: func(file string) error {
			contents, filename, err := loadStylesheet(file)
			if err != nil {
				return err
			}
			updateMap(func(styles map[string]string) {
				styles[filename] = contents
			})
			return nil
		},
		clean: func(file string) error {
			updateMap(func(styles map[string]string) {
				delete(styles, file+".css")
			})
			return nil
		},
	}
	return ln
}

func loadStylesheet(filename string) (string, string, error) {
	if strings.HasSuffix(filename, ".scss") {
		return loadSCSS(filename)
	}
	return loadCSS(filename)
}

func loadSCSS(filename string) (string, string, error) {
	in, err := os.Open("styles/" + filename)
	if err != nil {
		return "", "", fmt.Errorf("could not open stylesheet %s: %w", filename, err)
	}
	stylesheet, err := io.ReadAll(in)
	if err != nil {
		return "", "", fmt.Errorf("could not read stylesheet %s: %w", filename, err)
	}
	if sassTranspiler == nil {
		sassTranspiler, err = godartsass.Start(godartsass.Options{})
		if err != nil {
			return "", "", fmt.Errorf("could not start sass transpiler: %w", err)
		}
	}
	res, err := sassTranspiler.Execute(godartsass.Args{
		Source: string(stylesheet),
	})
	if err != nil {
		return "", "", fmt.Errorf("could not generate stylesheet %s: %w", filename, err)
	}
	return res.CSS, strings.TrimSuffix(filename, ".scss") + ".css", nil
}

func loadCSS(filename string) (string, string, error) {
	in, err := os.Open("styles/" + filename)
	if err != nil {
		return "", "", fmt.Errorf("could not open style infile %s: %w", filename, err)
	}
	var buf strings.Builder
	_, err = io.Copy(&buf, in)
	if err != nil {
		return "", "", fmt.Errorf("could not copy stylesheet %s: %s", filename, err)
	}
	return buf.String(), filename, nil
}
