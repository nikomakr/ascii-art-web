package main

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type PageData struct { // PageData struct to hold data for the template
	Result string
	Text   string
	Banner string
	Error string
}

var templates *template.Template // Global variable to hold parsed templates

func main() {
	var err error
	templates, err = template.ParseGlob("templates/*.html") // Parse all HTML templates in the templates directory. ParseGlob is used to parse multiple templates at once, and the pattern "templates/*.html" tells it to look for all HTML files in the templates directory.
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ascii-art", asciiArtHandler)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil { // Start the server on port 8080. ListenAndServe takes the address to listen on and a handler (nil means use the default handler). If there is an error starting the server, it will log the error and exit.
		log.Fatalf("Server failed: %v", err)
	}
}