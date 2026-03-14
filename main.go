package main

import (
	"bufio"
	"encoding/json"
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
	Error  string
}

var templates *template.Template // Global variable to hold parsed templates
var bannerCache map[string]map[rune][]string // Cache of loaded banner charMaps
var validBanners = map[string]bool{"standard": true, "shadow": true, "thinkertoy": true}

func parseTemplates() (*template.Template, error) {
	return template.ParseGlob("templates/*.html")
}

func writeJSON(w http.ResponseWriter, code int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func hasNonASCII(text string) bool {
	for _, ch := range text {
		if ch != '\n' && (ch < 32 || ch > 126) {
			return true
		}
	}
	return false
}

func main() {
	var err error
	templates, err = parseTemplates()
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	bannerCache = make(map[string]map[rune][]string)
	for _, name := range []string{"standard", "shadow", "thinkertoy"} {
		charMap, err := loadBanner("banners/" + name + ".txt")
		if err != nil {
			log.Fatalf("Error loading banner %s: %v", name, err)
		}
		bannerCache[name] = charMap
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ascii-art", asciiArtHandler)
	http.HandleFunc("/api/ascii-art", apiAsciiArtHandler)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil { // Start the server on port 8080. ListenAndServe takes the address to listen on and a handler (nil means use the default handler). If there is an error starting the server, it will log the error and exit.
		log.Fatalf("Server failed: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		renderError(w, http.StatusNotFound, "404 Not Found")
		return
	}
	if r.Method != http.MethodGet {
		renderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}
	data := PageData{Banner: "standard"}
	renderTemplate(w, "index.html", data)
}

func asciiArtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderError(w, http.StatusMethodNotAllowed, "405 Method Not Allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		renderError(w, http.StatusBadRequest, "400 Bad Request: could not parse form")
		return
	}

	text := r.FormValue("text")
	banner := r.FormValue("banner")

	if !validBanners[banner] {
		renderError(w, http.StatusBadRequest, "400 Bad Request: invalid banner")
		return
	}

	if hasNonASCII(text) {
		renderError(w, http.StatusBadRequest, "400 Bad Request: non-ASCII character")
		return
	}

	result, err := generateASCIIArt(text, banner)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "500 Internal Server Error: "+err.Error())
		return
	}

	data := PageData{Result: result, Text: text, Banner: banner}
	renderTemplate(w, "index.html", data)
}

func apiAsciiArtHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "405 Method Not Allowed"})
		return
	}

	var req struct {
		Text   string `json:"text"`
		Banner string `json:"banner"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "400 Bad Request: invalid JSON"})
		return
	}

	if !validBanners[req.Banner] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "400 Bad Request: invalid banner"})
		return
	}

	if hasNonASCII(req.Text) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "400 Bad Request: non-ASCII character"})
		return
	}

	result, err := generateASCIIArt(req.Text, req.Banner)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "500 Internal Server Error: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"result": result})
}

func generateASCIIArt(text, bannerName string) (string, error) {
	charMap := bannerCache[bannerName]

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
		renderError(w, http.StatusNotFound, "404 Not Found")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
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
