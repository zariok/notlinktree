package main

import (
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	w      http.ResponseWriter
	status int
}

func (r *statusRecorder) Header() http.Header         { return r.w.Header() }
func (r *statusRecorder) Write(b []byte) (int, error) { return r.w.Write(b) }
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.w.WriteHeader(status)
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{w: w, status: 200}
		start := time.Now()
		next.ServeHTTP(rec, r)
		dur := time.Since(start)
		log.Printf("%s | %s %s %d %s", start.Format("2006/01/02 - 15:04:05"), r.Method, r.URL.Path, rec.status, dur)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
