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

func generateASCIIArt(text, bannerName string) (string, error) {
	charMap, err := loadBanner("banners/" + bannerName + ".txt")
	if err != nil {
		return "", err
	}

	if text == "" {
		return "", nil
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if line == "" {
			result.WriteString("\n")
			continue
		}
		for row := 0; row < 8; row++ {
			for _, ch := range line {
				charLines, ok := charMap[ch]
				if !ok {
					charLines = charMap[' ']
				}
				result.WriteString(charLines[row])
			}
			result.WriteString("\n")
		}
	}
	return result.String(), nil
}

func loadBanner(path string) (map[rune][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	charMap := make(map[rune][]string)
	for code := 32; code <= 126; code++ {
		start := (code - 32) * 9
		if start+8 > len(allLines) {
			break
		}
		charMap[rune(code)] = allLines[start : start+8]
	}
	return charMap, nil
}

func renderTemplate(w http.ResponseWriter, tmplName string, data PageData) {
	t := templates.Lookup(tmplName)
	if t == nil {
		renderError(w, http.StatusNotFound, "404 - Template Not Found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		log.Printf("Template error: %v", err)
	}
}

func renderError(w http.ResponseWriter, code int, message string) {
	t := templates.Lookup("error.html")
	if t == nil {
		http.Error(w, message, code)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	if err := t.Execute(w, PageData{Error: message}); err != nil {
		http.Error(w, message, code)
	}
}