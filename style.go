package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/wellington/go-libsass"
)

func newStylesMap() (map[string]string, error) {
	folder, err := os.ReadDir("styles/")
	if err != nil {
		return nil, fmt.Errorf("Could not load styles directory: %s", err)
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
		return "", "", fmt.Errorf("Could not open style infile %s: %w", filename, err)
	}
	var buf strings.Builder
	comp, err := libsass.New(&buf, in)
	if err != nil {
		return "", "", fmt.Errorf("Could not start sass compiler for file %s: %w", filename, err)
	}
	if err = comp.Run(); err != nil {
		return "", "", fmt.Errorf("Could not generate stylesheet %s: %w", filename, err)
	}
	return buf.String(), strings.TrimSuffix(filename, ".scss") + ".css", nil
}

func loadCSS(filename string) (string, string, error) {
	in, err := os.Open("styles/" + filename)
	if err != nil {
		return "", "", fmt.Errorf("Could not open style infile %s: %w", filename, err)
	}
	var buf strings.Builder
	_, err = io.Copy(&buf, in)
	if err != nil {
		return "", "", fmt.Errorf("Could not copy stylesheet %s: %s", filename, err)
	}
	return buf.String(), filename, nil
}
