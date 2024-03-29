package main

import (
	"log"
	"net/http"
)

func main() {
	log.Printf("Hello, world! This is Prose.")

	s, err := newServer()

	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", s.router)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
