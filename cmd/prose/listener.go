package main

import (
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/rjeczalik/notify"
)

type listener struct {
	folder string
	update func(string) error
	clean  func(string) error
}

func (l *listener) listen() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("could not get current working directory for listener!")
	}
	cwd = cwd + "/"

	c := make(chan notify.EventInfo, 1)

	var events []notify.Event

	// inotify events prevent double-firing of
	// certain events in Linux.
	if runtime.GOOS == "linux" {
		events = []notify.Event{
			notify.InCloseWrite,
			notify.InMovedFrom,
			notify.InMovedTo,
			notify.InDelete,
		}
	} else {
		events = []notify.Event{
			notify.Create,
			notify.Remove,
			notify.Rename,
			notify.Write,
		}
	}

	err = notify.Watch(l.folder, c, events...)

	if err != nil {
		log.Fatalf("Could not setup watcher for folder %s: %s", l.folder, err)
	}

	defer notify.Stop(c)

	for {
		ei := <-c
		log.Printf("event: %s", ei.Event())
		switch ei.Event() {
		case notify.InCloseWrite, notify.InMovedTo, notify.Create, notify.Rename, notify.Write:
			filePath := strings.TrimPrefix(ei.Path(), cwd)
			log.Printf("updating file %s", filePath)
			err := l.update(strings.TrimPrefix(filePath, l.folder))
			if err != nil {
				log.Printf("watcher update action on %s failed: %v", filePath, err)
			}
		case notify.InMovedFrom, notify.InDelete, notify.Remove:
			filePath := strings.TrimPrefix(ei.Path(), cwd)
			log.Printf("cleaning file %s", filePath)
			err := l.clean(strings.TrimPrefix(filePath, l.folder))
			if err != nil {
				log.Printf("watcher clean action on %s failed: %v", filePath, err)
			}
		}
	}
}
