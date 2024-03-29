package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aymerick/raymond"
)

func loadTemplate(file string) (*raymond.Template, error) {
	tpl, err := raymond.ParseFile("templates/" + file)
	if err != nil {
		return nil, fmt.Errorf("could not parse %s template: %w", file, err)
	}
	tpl.RegisterHelper("datetime", func(timeStr string) string {
		timestamp, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			log.Printf("Could not parse timestamp '%v', falling back to current time", timeStr)
			timestamp = time.Now().Unix()
		}
		return time.Unix(timestamp, 0).Format("Jan 2 2006, 3:04 PM")
	})
	tpl.RegisterHelper("rssDatetime", func(timeStr string) string {
		timestamp, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			log.Printf("Could not parse timestamp '%v', falling back to current time", timeStr)
			timestamp = time.Now().Unix()
		}
		return rssDatetime(timestamp)
	})
	tpl.RegisterHelper("getFullUrl", func(slug string) string {
		return blogURL + "/" + slug
	})
	log.Printf("Loaded template: %s", file)
	return tpl, nil
}

// loadTemplates, for each f in files, loads `templates/$f`
// as a handlebars HTML/XML template. If any single template fails to
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
			newTpl, err := loadTemplate(file)
			if err != nil {
				return err
			}
			update(func(oldMap map[string]*raymond.Template) {
				oldMap[file] = newTpl
			})
			return nil
		},
		clean: func(file string) error {
			update(func(oldMap map[string]*raymond.Template) {
				delete(oldMap, file)
			})
			log.Printf("Unloaded template: %s", file)
			return nil
		},
	}
}
