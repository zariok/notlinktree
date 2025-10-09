package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"embed"
)

var (
	config      Config
	configMu    sync.RWMutex
	jwtSecret   []byte
	rateLimiter = NewRateLimiter(1*time.Hour, 10) // 10 clicks per hour per IP
	configPath  string

	version   = "v0.0.0-dev"
	buildDate = "unknown"

	// New variables for click counts
	clickCounts = make(map[string]int)
	clicksMu    sync.Mutex
)

//go:embed embed/ui/*
var uiFS embed.FS

// isLocalhost checks if the given IP address is localhost
func isLocalhost(ip string) bool {
	// Remove port if present
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	// Check for common localhost addresses
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
}

// reloadRunningInstance attempts to reload the config in a running instance
func reloadRunningInstance() error {
	// Get port from environment variable
	port := os.Getenv("NLT_PORT")
	if port == "" {
		port = "8080"
	}

	// Construct the URL
	url := fmt.Sprintf("http://localhost:%s/api/refresh-config", port)

	// Make HTTP request
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to running instance: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("running instance returned status %d", resp.StatusCode)
	}

	return nil
}

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	showHelp := flag.Bool("help", false, "Show help and exit")
	setAdminPW := flag.String("setadminpw", "", "Set the admin password and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n  NLT_PORT   Set the port for the server (default: 8080)\n  NLT_DATA   Set the directory for config.yaml (default: .)\n")
	}
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if *showVersion {
		fmt.Printf("notlinktree %s\nBuilt: %s\n", version, buildDate)
		os.Exit(0)
	}

	// Use NLT_DATA for config directory, default to "."
	dataDir := os.Getenv("NLT_DATA")
	if dataDir == "" {
		dataDir = "."
	}
	configPath = filepath.Join(dataDir, "config.yaml")

	// Handle -setadminpw flag (before loading config with JWT secret)
	setAdminPWProvided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "setadminpw" {
			setAdminPWProvided = true
		}
	})

	if setAdminPWProvided {
		if err := saveAdminPassword(*setAdminPW, configPath); err != nil {
			log.Fatalf("Failed to set admin password: %v", err)
		}
		fmt.Printf("Admin password has been set successfully.\n")

		// Try to reload config in running instance
		if err := reloadRunningInstance(); err != nil {
			fmt.Printf("Warning: Could not reload config in running instance: %v\n", err)
			fmt.Printf("You may need to manually reload the config or restart the server.\n")
		} else {
			fmt.Printf("Config reloaded in running instance.\n")
		}
		os.Exit(0)
	}

	var err error
	config, err = loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	jwtSecret = []byte(os.Getenv("NLT_JWT_SECRET"))
	if len(jwtSecret) == 0 {
		log.Fatal("NLT_JWT_SECRET environment variable not set")
	}

	// Initialize click counts from config
	clicksMu.Lock()
	for id, link := range config.Links {
		clickCounts[id] = link.Clicks
	}
	clicksMu.Unlock()

	// Start background goroutine to flush clicks
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			flushClicks()
		}
	}()

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/links", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowedError(w, r.Method)
			return
		}
		configMu.RLock()
		linksSlice := make([]Link, 0, len(config.Links))
		for _, link := range config.Links {
			linksSlice = append(linksSlice, link)
		}
		configMu.RUnlock()
		writeJSONSuccess(w, map[string]interface{}{"links": linksSlice})
	})
	mux.HandleFunc("/api/click/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeMethodNotAllowedError(w, r.Method)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/click/")
		trackClickHTTP(w, r, id)
	})
	mux.HandleFunc("/api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeMethodNotAllowedError(w, r.Method)
			return
		}
		adminLoginHTTP(w, r, config)
	})
	mux.HandleFunc("/api/admin/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if !checkAuth(r) {
				writeUnauthorizedError(w, "Authentication required")
				return
			}
			configMu.RLock()
			defer configMu.RUnlock()

			// Create a copy to avoid modifying the original config
			configCopy := config
			configCopy.Admin.Password = "" // Never send the password

			// Convert map to slice for the frontend
			linksSlice := make([]Link, 0, len(configCopy.Links))
			for _, link := range configCopy.Links {
				linksSlice = append(linksSlice, link)
			}

			writeJSONSuccess(w, map[string]interface{}{
				"links": linksSlice,
				"ui":    configCopy.UI,
			})
			return
		}
		if r.Method == http.MethodPost {
			if !checkAuth(r) {
				writeUnauthorizedError(w, "Authentication required")
				return
			}
			var req struct {
				UI UIConfig `json:"ui"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeInvalidRequestError(w, "Invalid request format", "Please ensure the request contains valid JSON")
				return
			}
			configMu.Lock()
			config.UI = req.UI
			err := saveConfig(config, configPath)
			configMu.Unlock()
			if err != nil {
				log.Printf("admin config update: saveConfig error: %v", err)
				writeInternalServerError(w, "Unable to save configuration", "Please try again in a moment")
				return
			}
			writeJSONSuccess(w, map[string]string{"status": "UI config saved"})
			return
		}
		writeMethodNotAllowedError(w, r.Method)
	})
	mux.HandleFunc("/api/admin/links", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		if r.Method == http.MethodPost {
			addLinkHTTP(w, r)
			return
		}
		writeMethodNotAllowedError(w, r.Method)
	})
	mux.HandleFunc("/api/admin/links/", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/admin/links/")
		if r.Method == http.MethodPut {
			updateLinkHTTP(w, r, id)
			return
		} else if r.Method == http.MethodDelete {
			deleteLinkHTTP(w, r, id)
			return
		}
		writeMethodNotAllowedError(w, r.Method)
	})
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowedError(w, r.Method)
			return
		}
		writeJSONSuccess(w, config.UI)
	})
	mux.HandleFunc("/api/admin/refresh-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeMethodNotAllowedError(w, r.Method)
			return
		}
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		newConfig, err := loadConfig(configPath)
		if err != nil {
			log.Printf("admin refresh-config: loadConfig error: %v", err)
			writeJSONError(w, http.StatusInternalServerError, ErrCodeConfigReloadFailed,
				"Unable to reload configuration", "Please check the config file and try again")
			return
		}
		configMu.Lock()
		config = newConfig
		configMu.Unlock()

		writeJSONSuccess(w, map[string]string{"status": "Config refreshed from disk"})
	})
	mux.HandleFunc("/api/refresh-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeMethodNotAllowedError(w, r.Method)
			return
		}

		// Only allow localhost access
		clientIP := r.RemoteAddr
		if r.Header.Get("X-Forwarded-For") != "" {
			clientIP = r.Header.Get("X-Forwarded-For")
		}
		if r.Header.Get("X-Real-IP") != "" {
			clientIP = r.Header.Get("X-Real-IP")
		}

		// Check if the request is from localhost
		if !isLocalhost(clientIP) {
			writeUnauthorizedError(w, "Access denied: localhost only")
			return
		}

		newConfig, err := loadConfig(configPath)
		if err != nil {
			log.Printf("public refresh-config: loadConfig error: %v", err)
			writeJSONError(w, http.StatusInternalServerError, ErrCodeConfigReloadFailed,
				"Unable to reload configuration", "Please check the config file and try again")
			return
		}
		configMu.Lock()
		config = newConfig
		configMu.Unlock()

		writeJSONSuccess(w, map[string]string{"status": "Config refreshed from disk"})
	})
	mux.HandleFunc("/api/admin/avatar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAvatarGetHTTP(w, r)
			return
		}
		// Auth check for POST is in adminAvatarUploadHTTP
		adminAvatarUploadHTTP(w, r)
	})
	mux.HandleFunc("/api/avatar", avatarGetHTTP)
	mux.HandleFunc("/api/admin/password", func(w http.ResponseWriter, r *http.Request) {
		adminPasswordChangeHTTP(w, r)
	})

	// Unified SPA static handler for all non-API routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeNotFoundError(w, "API endpoint")
			return
		}
		spaHandler(uiFS, "embed/ui").ServeHTTP(w, r)
	})

	// Logging middleware (prints method, path, status)
	handler := withLogging(mux)
	// CORS middleware (simple, allow all)
	handler = withCORS(handler)

	port := os.Getenv("NLT_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
