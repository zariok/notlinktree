package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func trackClickHTTP(w http.ResponseWriter, r *http.Request, id string) {
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		ip = ip[:i]
	}
	if !rateLimiter.Allow(ip) {
		writeJSONError(w, http.StatusTooManyRequests, ErrCodeRateLimited,
			"Too many requests", "Please wait before clicking again")
		return
	}
	if id == "" {
		writeInvalidRequestError(w, "Link ID is required", "Please provide a valid link ID")
		return
	}

	configMu.RLock()
	_, ok := config.Links[id]
	configMu.RUnlock()

	if !ok {
		writeNotFoundError(w, "Link")
		return
	}

	clicksMu.Lock()
	clickCounts[id]++
	currentClicks := clickCounts[id]
	clicksMu.Unlock()

	writeJSONSuccess(w, map[string]interface{}{"clicks": currentClicks})
}

func adminLoginHTTP(w http.ResponseWriter, r *http.Request, config Config) {
	var login struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		writeInvalidRequestError(w, "Invalid request format", "Please ensure the request contains valid JSON with a password field")
		return
	}

	if login.Password == "" {
		writeInvalidRequestError(w, "Password is required", "Please provide a password")
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(login.Password))
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, ErrCodeInvalidCredentials,
			"Invalid password", "Please check your password and try again")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("admin login: token generation error: %v", err)
		writeInternalServerError(w, "Unable to create session", "Please try again in a moment")
		return
	}
	writeJSONSuccess(w, map[string]string{"token": tokenString})
}

func addLinkHTTP(w http.ResponseWriter, r *http.Request) {
	var newLink Link
	if err := json.NewDecoder(r.Body).Decode(&newLink); err != nil {
		writeInvalidRequestError(w, "Invalid link data", "Please ensure the request contains valid JSON with link information")
		return
	}

	// Validate required fields
	if newLink.Title == "" {
		writeInvalidRequestError(w, "Link title is required", "Please provide a title for your link")
		return
	}
	if newLink.URL == "" {
		writeInvalidRequestError(w, "Link URL is required", "Please provide a valid URL for your link")
		return
	}

	if newLink.ID == "" {
		newLink.ID = generateID()
	}
	configMu.Lock()
	defer configMu.Unlock()

	// If link exists, preserve click count
	if existingLink, ok := config.Links[newLink.ID]; ok {
		newLink.Clicks = existingLink.Clicks
	}

	config.Links[newLink.ID] = newLink
	err := saveConfig(config, configPath)
	if err != nil {
		log.Printf("addLinkHTTP: saveConfig error: %v", err)
		writeInternalServerError(w, "Unable to save link", "Please try again in a moment")
		return
	}
	writeJSONSuccess(w, newLink)
}

func updateLinkHTTP(w http.ResponseWriter, r *http.Request, id string) {
	var updatedFields Link
	if err := json.NewDecoder(r.Body).Decode(&updatedFields); err != nil {
		writeInvalidRequestError(w, "Invalid link data", "Please ensure the request contains valid JSON with link information")
		return
	}
	configMu.Lock()
	defer configMu.Unlock()

	link, ok := config.Links[id]
	if !ok {
		writeNotFoundError(w, "Link")
		return
	}

	// Merge fields
	if updatedFields.Title != "" {
		link.Title = updatedFields.Title
	}
	if updatedFields.URL != "" {
		link.URL = updatedFields.URL
	}
	if updatedFields.Type != "" {
		link.Type = updatedFields.Type
	}
	if updatedFields.Description != "" {
		link.Description = updatedFields.Description
	}
	config.Links[id] = link
	if err := saveConfig(config, configPath); err != nil {
		log.Printf("updateLinkHTTP: saveConfig error: %v", err)
		writeInternalServerError(w, "Unable to save link changes", "Please try again in a moment")
		return
	}
	writeJSONSuccess(w, link)
}

func deleteLinkHTTP(w http.ResponseWriter, r *http.Request, id string) {
	configMu.Lock()
	defer configMu.Unlock()

	if _, ok := config.Links[id]; !ok {
		writeNotFoundError(w, "Link")
		return
	}

	delete(config.Links, id)

	// also delete from clickCounts
	clicksMu.Lock()
	delete(clickCounts, id)
	clicksMu.Unlock()

	if err := saveConfig(config, configPath); err != nil {
		log.Printf("deleteLinkHTTP: saveConfig error: %v", err)
		writeInternalServerError(w, "Unable to delete link", "Please try again in a moment")
		return
	}
	writeJSONSuccess(w, map[string]string{"status": "success"})
}

// Helper to get the NLT_DATA directory
func getDataDir() string {
	dir := os.Getenv("NLT_DATA")
	if dir == "" {
		dir = "."
	}
	return dir
}

// POST /api/admin/avatar
func adminAvatarUploadHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		writeUnauthorizedError(w, "Authentication required")
		return
	}
	if r.Method != http.MethodPost {
		writeMethodNotAllowedError(w, r.Method)
		return
	}

	// Max 1MB
	r.ParseMultipartForm(1 << 20)
	file, _, err := r.FormFile("avatar")
	if err != nil {
		writeInvalidRequestError(w, "Invalid file upload", "Please select a valid image file")
		return
	}
	defer file.Close()

	filePath := filepath.Join(getDataDir(), "avatar.png")
	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("admin avatar upload: file creation error: %v", err)
		writeInternalServerError(w, "Unable to save avatar", "Please try again in a moment")
		return
	}
	defer dst.Close()

	io.Copy(dst, file)
	writeJSONSuccess(w, map[string]string{"status": "Avatar uploaded"})
}

// GET /api/admin/avatar
func adminAvatarGetHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		writeUnauthorizedError(w, "Authentication required")
		return
	}
	avatarGetHTTP(w, r)
}

// GET /api/avatar
func avatarGetHTTP(w http.ResponseWriter, r *http.Request) {
	avatarPath := filepath.Join(getDataDir(), "avatar.png")
	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		writeJSONError(w, http.StatusNotFound, ErrCodeAvatarNotFound,
			"Avatar not found", "No profile picture has been uploaded yet")
		return
	}
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "image/png")
	http.ServeFile(w, r, avatarPath)
}

func flushClicks() {
	clicksMu.Lock()
	if len(clickCounts) == 0 {
		clicksMu.Unlock()
		return
	}
	// Create a copy of the current click counts to avoid holding the lock for too long.
	countsSnapshot := make(map[string]int)
	for id, count := range clickCounts {
		countsSnapshot[id] = count
	}
	clicksMu.Unlock()

	configMu.Lock()
	defer configMu.Unlock()

	var hasChanges bool
	for id, currentClicks := range countsSnapshot {
		if link, ok := config.Links[id]; ok {
			if link.Clicks != currentClicks {
				link.Clicks = currentClicks
				config.Links[id] = link
				hasChanges = true
			}
		}
	}

	if hasChanges {
		if err := saveConfig(config, configPath); err != nil {
			log.Printf("Error flushing clicks to config: %v", err)
		} else {
			log.Println("Flushed link click counts to config.")
		}
	}
}

// POST /api/admin/password
func adminPasswordChangeHTTP(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		writeUnauthorizedError(w, "Authentication required")
		return
	}
	if r.Method != http.MethodPost {
		writeMethodNotAllowedError(w, r.Method)
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequestError(w, "Invalid request format", "Please ensure the request contains valid JSON with a password field")
		return
	}
	if req.Password == "" {
		writeInvalidRequestError(w, "Password is required", "Please provide a new password")
		return
	}

	// Use the reusable password save function
	err := saveAdminPassword(req.Password, configPath)
	if err != nil {
		log.Printf("admin password change: saveAdminPassword error: %v", err)
		writeInternalServerError(w, "Unable to save password", "Please try again in a moment")
		return
	}

	// Update the in-memory config to reflect the change
	configMu.Lock()
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		configMu.Unlock()
		log.Printf("admin password change: hash generation error: %v", err)
		writeInternalServerError(w, "Unable to process password", "Please try again in a moment")
		return
	}
	config.Admin.Password = string(hash)
	configMu.Unlock()
	writeJSONSuccess(w, map[string]string{"status": "Password changed"})
}
