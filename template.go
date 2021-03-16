package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aymerick/raymond"
)

func loadTemplate(file string) (*raymond.Template, error) {
	tpl, err := raymond.ParseFile("templates/" + file + ".html")
	if err != nil {
		return nil, fmt.Errorf("Could not parse %s template: %w", file, err)
	}
	tpl.RegisterHelper("datetime", func(timeStr string) string {
		timestamp, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			log.Printf("Could not parse timestamp '%v', falling back to current time", timeStr)
			timestamp = time.Now().Unix()
		}
		return time.Unix(timestamp, 0).Format("Jan 2 2006, 3:04 PM")
	})
	log.Printf("Loaded template: %s", file)
	return tpl, nil
}

// loadTemplates, for each f in files, loads `templates/$f.html`
// as a handlebars HTML template. If any single template fails to
// load, only an error is returned. Conversely, if there is no error,
// every template name passed is guaranteed to have loaded successfully.
func loadTemplates(files []string) (map[string]*raymond.Template, error) {
	templates := make(map[string]*raymond.Template)
	for _, f := range files {
		tpl, err := loadTemplate(f)
		if err != nil {
			return nil, err
		}
		templates[f] = tpl
	}
	log.Printf("Loaded templates: %s", files)
	return templates, nil
}

func newTemplateListener(update func(func(map[string]*raymond.Template))) *listener {
	return &listener{
		folder: "templates/",
		update: func(file string) error {
			tplName := strings.TrimSuffix(file, ".html")
			newTpl, err := loadTemplate(tplName)
			if err != nil {
				return err
			}
			update(func(oldMap map[string]*raymond.Template) {
				oldMap[tplName] = newTpl
			})
			return nil
		},
		clean: func(file string) error {
			tplName := strings.TrimSuffix(file, ".html")
			update(func(oldMap map[string]*raymond.Template) {
				delete(oldMap, tplName)
			})
			log.Printf("Unloaded template: %s", tplName)
			return nil
		},
	}
}
