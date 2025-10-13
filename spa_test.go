package main

import (
	"embed"
	"net/http/httptest"
	"strings"
	"testing"
)

//go:embed test/data/index.html
var testFS embed.FS

func TestGetContentType(t *testing.T) {
	cases := []struct {
		path string
		exp  string
	}{
		{"foo.css", "text/css"},
		{"foo.js", "application/javascript"},
		{"foo.json", "application/json"},
		{"foo.png", "image/png"},
		{"foo.jpg", "image/jpeg"},
		{"foo.gif", "image/gif"},
		{"foo.svg", "image/svg+xml"},
		{"foo.ico", "image/x-icon"},
		{"foo.html", "text/html"},
		{"foo.txt", "text/plain"},
	}
	for _, c := range cases {
		if got := getContentType(c.path); got != c.exp {
			t.Errorf("getContentType(%q) = %q, want %q", c.path, got, c.exp)
		}
	}
}

func TestSPAHandler_ServesIndex(t *testing.T) {
	h := spaHandler(testFS, "test/data")

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if !strings.Contains(rw.Body.String(), "<html>") {
		t.Errorf("Expected index.html content, got: %s", rw.Body.String())
	}
	if ct := rw.Header().Get("Content-Type"); ct != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", ct)
	}
}

func TestSPAHandler_MissingFile(t *testing.T) {
	h := spaHandler(testFS, "test/data")
	req := httptest.NewRequest("GET", "/notfound.js", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 200 || !strings.Contains(rw.Body.String(), "<html>") {
		t.Errorf("Expected fallback to index.html for missing file, got code %d", rw.Code)
	}
}

func TestSPAHandler_PathTraversal(t *testing.T) {
	h := spaHandler(testFS, "test/data")
	req := httptest.NewRequest("GET", "/../secret.txt", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 200 || !strings.Contains(rw.Body.String(), "<html>") {
		t.Errorf("Expected fallback to index.html for path traversal, got code %d", rw.Code)
	}
}
