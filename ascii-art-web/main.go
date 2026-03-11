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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "404 - Page Not Found")
		return
	}
	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "405 - Method Not Allowed")
		return
	}
	data := PageData{Banner: "standard"}
	renderTemplate(w, "index.html", data)
}

func asciiArtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderError(w, http.StatusMethodNotAllowed, "405 - Method Not Allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		renderError(w, http.StatusBadRequest, "400 - Bad Request: could not parse form")
		return
	}

	text := r.FormValue("text")
	banner := r.FormValue("banner")

	validBanners := map[string]bool{"standard": true, "shadow": true, "thinkertoy": true}
	if !validBanners[banner] {
		renderError(w, http.StatusBadRequest, "400 - Bad Request: invalid banner")
		return
	}

	for _, ch := range text {
		if ch != '\n' && (ch < 32 || ch > 126) {
			renderError(w, http.StatusBadRequest, "400 - Bad Request: non-ASCII character")
			return
		}
	}

	result, err := generateASCIIArt(text, banner)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "500 - Internal Server Error: "+err.Error())
		return
	}

	data := PageData{Result: result, Text: text, Banner: banner}
	renderTemplate(w, "index.html", data)
}