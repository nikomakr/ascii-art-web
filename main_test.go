package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func init() {
	var err error
	templates, err = parseTemplates()
	if err != nil {
		panic("failed to parse templates: " + err.Error())
	}

	bannerCache = make(map[string]map[rune][]string)
	for _, name := range []string{"standard", "shadow", "thinkertoy"} {
		charMap, err := loadBanner("banners/" + name + ".txt")
		if err != nil {
			panic("failed to load banner " + name + ": " + err.Error())
		}
		bannerCache[name] = charMap
	}
}

func TestHomeHandler_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	homeHandler(w, req)
	if w.Result().Status != "200 OK" {
		t.Errorf("expected 200 OK, got %s", w.Result().Status)
	}
}

func TestHomeHandler_NotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	homeHandler(w, req)
	if w.Result().Status != "404 Not Found" {
		t.Errorf("expected 404 Not Found, got %s", w.Result().Status)
	}
}

func TestHomeHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	homeHandler(w, req)
	if w.Result().Status != "405 Method Not Allowed" {
		t.Errorf("expected 405 Method Not Allowed, got %s", w.Result().Status)
	}
}

func TestAsciiArtHandler_ValidInput(t *testing.T) {
	form := url.Values{"text": {"Hello"}, "banner": {"standard"}}
	req := httptest.NewRequest(http.MethodPost, "/ascii-art", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	asciiArtHandler(w, req)
	if w.Result().Status != "200 OK" {
		t.Errorf("expected 200 OK, got %s", w.Result().Status)
	}
}

func TestAsciiArtHandler_InvalidBanner(t *testing.T) {
	form := url.Values{"text": {"Hi"}, "banner": {"unknown"}}
	req := httptest.NewRequest(http.MethodPost, "/ascii-art", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	asciiArtHandler(w, req)
	if w.Result().Status != "400 Bad Request" {
		t.Errorf("expected 400 Bad Request, got %s", w.Result().Status)
	}
}

func TestAsciiArtHandler_NonASCII(t *testing.T) {
	form := url.Values{"text": {"héllo"}, "banner": {"standard"}}
	req := httptest.NewRequest(http.MethodPost, "/ascii-art", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	asciiArtHandler(w, req)
	if w.Result().Status != "400 Bad Request" {
		t.Errorf("expected 400 Bad Request, got %s", w.Result().Status)
	}
}

func TestAsciiArtHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ascii-art", nil)
	w := httptest.NewRecorder()
	asciiArtHandler(w, req)
	if w.Result().Status != "405 Method Not Allowed" {
		t.Errorf("expected 405 Method Not Allowed, got %s", w.Result().Status)
	}
}

func TestRenderError_InternalServerError(t *testing.T) {
	w := httptest.NewRecorder()
	renderError(w, http.StatusInternalServerError, "500 Internal Server Error")
	if w.Result().Status != "500 Internal Server Error" {
		t.Errorf("expected 500 Internal Server Error, got %s", w.Result().Status)
	}
}

func TestGenerateASCIIArt_EmptyText(t *testing.T) {
	result, err := generateASCIIArt("", "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}
}

func TestGenerateASCIIArt_Newline(t *testing.T) {
	result, err := generateASCIIArt("\n", "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "\n") {
		t.Error("expected newline in result")
	}
}
