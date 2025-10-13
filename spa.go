package main

import (
	"embed"
	"net/http"
	"strings"
)

func spaHandler(fs embed.FS, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the request path
		origPath := r.URL.Path

		// Build the file path by combining prefix with the request path
		filePath := prefix + origPath

		// If the path ends with /, add index.html
		if strings.HasSuffix(origPath, "/") {
			filePath += "index.html"
		}

		data, err := fs.ReadFile(filePath)
		if err != nil {
			// fallback to /admin/index.html for /admin routes, else /index.html
			fallback := "/index.html"
			if strings.HasPrefix(origPath, "/admin") {
				fallback = "/admin/index.html"
			}
			data, err = fs.ReadFile(prefix + fallback)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not found"))
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(data)
			return
		}

		// Set content type and cache headers for static assets
		ctype := getContentType(filePath)
		w.Header().Set("Content-Type", ctype)
		if ctype != "text/html" {
			// Cache static assets for 1 year
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			// No cache for HTML entry points
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		w.Write(data)
	})
}

func getContentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".css"):
		return "text/css"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript"
	case strings.HasSuffix(path, ".json"):
		return "application/json"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(path, ".gif"):
		return "image/gif"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".ico"):
		return "image/x-icon"
	case strings.HasSuffix(path, ".html"):
		return "text/html"
	default:
		return "text/plain"
	}
}
