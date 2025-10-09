package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithCORS(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	ts := httptest.NewServer(withCORS(h))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header not set")
	}
	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "GET") {
		t.Error("CORS methods header missing GET")
	}
}

func TestWithCORS_OPTIONS(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not see this"))
	})
	ts := httptest.NewServer(withCORS(h))
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OPTIONS failed: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected 204 No Content, got %d", resp.StatusCode)
	}
}

func TestWithLogging(t *testing.T) {
	logged := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	// Wrap with a logger that sets a flag
	logger := withLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logged = true
		h.ServeHTTP(w, r)
	}))
	ts := httptest.NewServer(logger)
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	if !logged {
		t.Error("Expected logging to occur")
	}
}

func TestWithLogging_ErrorStatus(t *testing.T) {
	var status int
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	logger := withLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w2 := &statusRecorder{w: w, status: 200}
		h.ServeHTTP(w2, r)
		status = w2.status
	}))
	ts := httptest.NewServer(logger)
	defer ts.Close()
	_, _ = http.Get(ts.URL)
	if status != http.StatusTeapot {
		t.Errorf("Expected status 418, got %d", status)
	}
}

func TestWithCORS_AllMethods(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	ts := httptest.NewServer(withCORS(h))
	defer ts.Close()
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	for _, m := range methods {
		req, _ := http.NewRequest(m, ts.URL, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("%s failed: %v", m, err)
			continue
		}
		if allow := resp.Header.Get("Access-Control-Allow-Methods"); !strings.Contains(allow, m) {
			t.Errorf("CORS methods header missing %s", m)
		}
	}
}
