package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

var serialTestMu sync.Mutex

func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	// Use a temporary directory for all test artifacts
	tempDir, err := os.MkdirTemp("", "notlinktree-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Set environment variables for the test
	os.Setenv("NLT_DATA", tempDir)
	os.Setenv("NLT_JWT_SECRET", "test-secret-for-e2e-tests")
	jwtSecret = []byte(os.Getenv("NLT_JWT_SECRET"))

	// Create a test config file inside the temp directory
	configPath = filepath.Join(tempDir, "config.yaml")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	testConfig := Config{
		Admin: struct {
			Password string `yaml:"password" json:"-"`
		}{
			Password: string(hashedPassword),
		},
		Links: map[string]Link{
			"testlink1": {ID: "testlink1", Title: "Test Link 1", URL: "https://example.com/1", Clicks: 5},
		},
		UI: UIConfig{
			Username: "Test User",
			Title:    "My Test Links",
		},
	}

	// Save the initial test config
	if err := saveConfig(testConfig, configPath); err != nil {
		t.Fatalf("Failed to save initial test config: %v", err)
	}

	// Load the config into the global variable used by handlers
	config, err = loadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Initialize click counts from the loaded config
	clicksMu.Lock()
	clickCounts = make(map[string]int)
	for id, link := range config.Links {
		clickCounts[id] = link.Clicks
	}
	clicksMu.Unlock()

	// This is a simplified version of the router from main.go
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links", func(w http.ResponseWriter, r *http.Request) {
		configMu.RLock()
		linksSlice := make([]Link, 0, len(config.Links))
		for _, link := range config.Links {
			linksSlice = append(linksSlice, link)
		}
		configMu.RUnlock()
		writeJSONSuccess(w, map[string]interface{}{"links": linksSlice})
	})
	mux.HandleFunc("/api/click/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/click/")
		trackClickHTTP(w, r, id)
	})
	mux.HandleFunc("/api/admin/login", func(w http.ResponseWriter, r *http.Request) {
		adminLoginHTTP(w, r, config)
	})
	mux.HandleFunc("/api/admin/config", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		configCopy := config
		configCopy.Admin.Password = "" // Redact password
		writeJSONSuccess(w, configCopy)
	})
	mux.HandleFunc("/api/admin/links", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		addLinkHTTP(w, r)
	})
	mux.HandleFunc("/api/admin/links/", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/admin/links/")
		if r.Method == http.MethodPut {
			updateLinkHTTP(w, r, id)
		} else if r.Method == http.MethodDelete {
			deleteLinkHTTP(w, r, id)
		} else {
			writeMethodNotAllowedError(w, r.Method)
		}
	})
	mux.HandleFunc("/api/admin/avatar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			adminAvatarGetHTTP(w, r)
		} else if r.Method == http.MethodPost {
			if !checkAuth(r) {
				writeUnauthorizedError(w, "Authentication required")
				return
			}
			adminAvatarUploadHTTP(w, r)
		}
	})
	mux.HandleFunc("/api/avatar", avatarGetHTTP)
	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		writeJSONSuccess(w, config.UI)
	})
	mux.HandleFunc("/api/admin/password", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			writeUnauthorizedError(w, "Authentication required")
			return
		}
		adminPasswordChangeHTTP(w, r)
	})
	mux.HandleFunc("/api/refresh-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeMethodNotAllowedError(w, r.Method)
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

	server := httptest.NewServer(mux)

	// Teardown function to clean up resources
	teardown := func() {
		server.Close()
		os.RemoveAll(tempDir)
		os.Unsetenv("NLT_DATA")
		os.Unsetenv("NLT_JWT_SECRET")
	}

	return server, teardown
}

func TestAPI_AdminLogin(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	// Correct password
	loginDetails := map[string]string{"password": "testpass"}
	body, _ := json.Marshal(loginDetails)
	res, err := http.Post(server.URL+"/api/admin/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", res.StatusCode)
	}
	var tokenRes APIResponse
	json.NewDecoder(res.Body).Decode(&tokenRes)
	if !tokenRes.Success || tokenRes.Data == nil {
		t.Error("Expected successful response with token data")
	}
	if data, ok := tokenRes.Data.(map[string]interface{}); !ok || data["token"] == "" {
		t.Error("Expected a token in response data")
	}

	// Incorrect password
	loginDetails["password"] = "wrongpass"
	body, _ = json.Marshal(loginDetails)
	res, _ = http.Post(server.URL+"/api/admin/login", "application/json", bytes.NewBuffer(body))
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 Unauthorized for wrong password, got %d", res.StatusCode)
	}
}

func getAuthToken(t *testing.T, serverURL string) string {
	loginDetails := map[string]string{"password": "testpass"}
	body, _ := json.Marshal(loginDetails)
	res, err := http.Post(serverURL+"/api/admin/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to get auth token: %v", err)
	}
	var tokenRes APIResponse
	json.NewDecoder(res.Body).Decode(&tokenRes)
	if !tokenRes.Success || tokenRes.Data == nil {
		t.Fatal("Failed to get a valid token for tests")
	}
	data, ok := tokenRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	return data["token"].(string)
}

func TestAPI_LinkManagement(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)
	authHeader := "Bearer " + token

	// 1. Add a new link
	newLink := Link{ID: "newlink2", Title: "New Link 2", URL: "https://example.com/2"}
	body, _ := json.Marshal(newLink)
	req, _ := http.NewRequest("POST", server.URL+"/api/admin/links", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Add link request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		t.Fatalf("Expected status 200 on add, got %d. Body: %s", res.StatusCode, string(bodyBytes))
	}

	// 2. Verify the new link exists via public endpoint
	res, _ = http.Get(server.URL + "/api/links")
	bodyBytes, _ := io.ReadAll(res.Body)
	var linksRes APIResponse
	json.Unmarshal(bodyBytes, &linksRes)
	if !linksRes.Success || linksRes.Data == nil {
		t.Fatalf("Expected successful response with links data, got: %+v, body: %s", linksRes, string(bodyBytes))
	}
	data, ok := linksRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	if links, ok := data["links"].([]interface{}); !ok || len(links) != 2 {
		t.Fatalf("Expected 2 links after adding one, got %d", len(links))
	}

	// 3. Update the link
	updatedLink := Link{Title: "Updated Link 2 Title"}
	body, _ = json.Marshal(updatedLink)
	req, _ = http.NewRequest("PUT", server.URL+"/api/admin/links/newlink2", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 on update, got %d", res.StatusCode)
	}
	var updatedLinkRes APIResponse
	json.NewDecoder(res.Body).Decode(&updatedLinkRes)
	if !updatedLinkRes.Success || updatedLinkRes.Data == nil {
		t.Fatal("Expected successful response with link data")
	}
	// The data comes back as map[string]interface{} from JSON unmarshaling
	dataMap, ok := updatedLinkRes.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid response data format, got: %+v", updatedLinkRes.Data)
	}
	if title, ok := dataMap["title"].(string); !ok || title != "Updated Link 2 Title" {
		t.Errorf("Expected updated title 'Updated Link 2 Title', got '%s'", dataMap["title"])
	}

	// 4. Delete the link
	req, _ = http.NewRequest("DELETE", server.URL+"/api/admin/links/newlink2", nil)
	req.Header.Set("Authorization", authHeader)
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 on delete, got %d", res.StatusCode)
	}

	// 5. Verify it's gone
	res, _ = http.Get(server.URL + "/api/links")
	json.NewDecoder(res.Body).Decode(&linksRes)
	if !linksRes.Success || linksRes.Data == nil {
		t.Fatal("Expected successful response with links data")
	}
	linksData, ok := linksRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	if links, ok := linksData["links"].([]interface{}); !ok || len(links) != 1 {
		t.Errorf("Expected 1 link after deleting one, got %d", len(links))
	}
}

func TestAPI_TrackClick(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	// Initial clicks for testlink1 is 5
	res, err := http.Post(server.URL+"/api/click/testlink1", "application/json", nil)
	if err != nil {
		t.Fatalf("Click request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for click, got %d", res.StatusCode)
	}
	var clickRes APIResponse
	json.NewDecoder(res.Body).Decode(&clickRes)
	if !clickRes.Success || clickRes.Data == nil {
		t.Fatal("Expected successful response with click data")
	}
	data, ok := clickRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	// The in-memory count should now be 6
	if clicks, ok := data["clicks"].(float64); !ok || clicks != 6 {
		t.Errorf("Expected 6 clicks, got %v", data["clicks"])
	}

	// Flush clicks to disk to test persistence (not ideal, but for test completeness)
	flushClicks()

	// Load config again and check if clicks were persisted
	reloadedConfig, _ := loadConfig(configPath)
	if reloadedConfig.Links["testlink1"].Clicks != 6 {
		t.Errorf("Expected persisted click count to be 6, got %d", reloadedConfig.Links["testlink1"].Clicks)
	}
}

func TestAPI_Avatar(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)

	// 1. Check that avatar doesn't exist initially
	res, _ := http.Get(server.URL + "/api/avatar")
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for missing avatar, got %d", res.StatusCode)
	}
	req, _ := http.NewRequest("GET", server.URL+"/api/admin/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for missing admin avatar, got %d", res.StatusCode)
	}

	// 2. Upload a new avatar
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("avatar", "test_avatar.png")
	fw.Write([]byte("fake-image-data"))
	w.Close()

	req, _ = http.NewRequest("POST", server.URL+"/api/admin/avatar", &b)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		t.Fatalf("Expected 200 on avatar upload, got %d. Body: %s", res.StatusCode, string(bodyBytes))
	}

	// 3. Verify avatar can be fetched from public and admin endpoints
	res, _ = http.Get(server.URL + "/api/avatar")
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for existing avatar, got %d", res.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(res.Body)
	if string(bodyBytes) != "fake-image-data" {
		t.Error("Avatar content mismatch")
	}

	req, _ = http.NewRequest("GET", server.URL+"/api/admin/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 for existing admin avatar, got %d", res.StatusCode)
	}
}

func TestAPI_Unauthorized(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/config"},
		{"POST", "/api/admin/links"},
		{"PUT", "/api/admin/links/some-id"},
		{"DELETE", "/api/admin/links/some-id"},
		{"POST", "/api/admin/avatar"},
		{"GET", "/api/admin/avatar"},
		{"POST", "/api/admin/password"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+ep.path, func(t *testing.T) {
			req, _ := http.NewRequest(ep.method, server.URL+ep.path, nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			res, _ := http.DefaultClient.Do(req)
			if res.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected status 401 Unauthorized for %s %s, got %d", ep.method, ep.path, res.StatusCode)
			}
		})
	}
}

func TestAPI_AdminPasswordChange(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)
	authHeader := "Bearer " + token

	// Test successful password change
	newPassword := "newtestpass123"
	passwordData := map[string]string{"password": newPassword}
	body, _ := json.Marshal(passwordData)

	req, _ := http.NewRequest("POST", server.URL+"/api/admin/password", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Password change request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		t.Fatalf("Expected status 200 on password change, got %d. Body: %s", res.StatusCode, string(bodyBytes))
	}

	var response APIResponse
	json.NewDecoder(res.Body).Decode(&response)
	if !response.Success || response.Data == nil {
		t.Fatal("Expected successful response with status data")
	}
	statusMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	if status, ok := statusMap["status"].(string); !ok || status != "Password changed" {
		t.Errorf("Expected status 'Password changed', got '%s'", statusMap["status"])
	}

	// Test that new password works
	loginDetails := map[string]string{"password": newPassword}
	loginBody, _ := json.Marshal(loginDetails)
	res, err = http.Post(server.URL+"/api/admin/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with new password, got %d", res.StatusCode)
	}

	// Test invalid request (empty password)
	passwordData["password"] = ""
	emptyBody, _ := json.Marshal(passwordData)
	req, _ = http.NewRequest("POST", server.URL+"/api/admin/password", bytes.NewBuffer(emptyBody))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty password, got %d", res.StatusCode)
	}

	// Test invalid request (malformed JSON)
	req, _ = http.NewRequest("POST", server.URL+"/api/admin/password", strings.NewReader("invalid json"))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for malformed JSON, got %d", res.StatusCode)
	}

	// Test wrong method (GET instead of POST)
	req, _ = http.NewRequest("GET", server.URL+"/api/admin/password", nil)
	req.Header.Set("Authorization", authHeader)

	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for wrong method, got %d", res.StatusCode)
	}
}

func TestAPI_LinkManagement_EdgeCases(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)
	authHeader := "Bearer " + token

	// Test adding link with invalid data
	invalidLink := map[string]interface{}{"invalid": "data"}
	body, _ := json.Marshal(invalidLink)
	req, _ := http.NewRequest("POST", server.URL+"/api/admin/links", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Adding invalid link returned status %d, expected 400", res.StatusCode)
	}

	// Test updating non-existent link
	updateData := map[string]string{"title": "Non-existent"}
	body, _ = json.Marshal(updateData)
	req, _ = http.NewRequest("PUT", server.URL+"/api/admin/links/nonexistent", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Updating non-existent link returned status %d, expected 404", res.StatusCode)
	}

	// Test deleting non-existent link
	req, _ = http.NewRequest("DELETE", server.URL+"/api/admin/links/nonexistent", nil)
	req.Header.Set("Authorization", authHeader)

	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Deleting non-existent link returned status %d, expected 404", res.StatusCode)
	}
}

func TestAPI_TrackClick_EdgeCases(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	// Test clicking on non-existent link
	res, err := http.Post(server.URL+"/api/click/nonexistent", "application/json", nil)
	if err != nil {
		t.Fatalf("Click request on non-existent link failed: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent link, got %d", res.StatusCode)
	}

	// Get current click count before rapid clicking
	initialRes, _ := http.Post(server.URL+"/api/click/testlink1", "application/json", nil)
	var initialClickRes APIResponse
	json.NewDecoder(initialRes.Body).Decode(&initialClickRes)
	if !initialClickRes.Success || initialClickRes.Data == nil {
		t.Fatal("Expected successful response with click data")
	}
	data, ok := initialClickRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	initialClicks := int(data["clicks"].(float64))

	// Test multiple rapid clicks
	for i := 0; i < 5; i++ {
		res, _ := http.Post(server.URL+"/api/click/testlink1", "application/json", nil)
		if res.StatusCode != http.StatusOK {
			t.Errorf("Rapid click %d failed with status %d", i+1, res.StatusCode)
		}
	}

	// Verify click count increased
	flushClicks()
	reloadedConfig, _ := loadConfig(configPath)
	expectedClicks := initialClicks + 5 // initial + rapid clicks
	if reloadedConfig.Links["testlink1"].Clicks != expectedClicks {
		t.Errorf("Expected %d clicks after rapid clicking, got %d", expectedClicks, reloadedConfig.Links["testlink1"].Clicks)
	}
}

func TestAPI_Avatar_EdgeCases(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)

	// Test uploading avatar without file
	req, _ := http.NewRequest("POST", server.URL+"/api/admin/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/form-data")

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Uploading avatar without file returned status %d, expected 400", res.StatusCode)
	}

	// Test uploading avatar with invalid content type
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, _ := w.CreateFormFile("avatar", "test.txt")
	fw.Write([]byte("not-an-image"))
	w.Close()

	req, _ = http.NewRequest("POST", server.URL+"/api/admin/avatar", &b)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Errorf("Uploading non-image file returned status %d, expected 200", res.StatusCode)
	}
}

func TestAPI_Config_EdgeCases(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	// Test config endpoint
	res, err := http.Get(server.URL + "/api/config")
	if err != nil {
		t.Fatalf("Config request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for config, got %d", res.StatusCode)
	}

	var configRes APIResponse
	json.NewDecoder(res.Body).Decode(&configRes)
	if !configRes.Success || configRes.Data == nil {
		t.Fatal("Expected successful response with config data")
	}
	// The data comes back as map[string]interface{} from JSON unmarshaling
	configMap, ok := configRes.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Invalid response data format, got: %+v", configRes.Data)
	}
	if username, ok := configMap["username"].(string); !ok || username != "Test User" {
		t.Errorf("Expected username 'Test User', got '%s'", configMap["username"])
	}

	// Test links endpoint
	res, err = http.Get(server.URL + "/api/links")
	if err != nil {
		t.Fatalf("Links request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for links, got %d", res.StatusCode)
	}

	var linksRes APIResponse
	json.NewDecoder(res.Body).Decode(&linksRes)
	if !linksRes.Success || linksRes.Data == nil {
		t.Fatal("Expected successful response with links data")
	}
	linksData, ok := linksRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	if links, ok := linksData["links"].([]interface{}); !ok || len(links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(links))
	}
}

func TestAPI_AdminPasswordChange_WithReusableFunction(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)
	authHeader := "Bearer " + token

	// Test successful password change using the new reusable function
	newPassword := "newtestpass456"
	passwordData := map[string]string{"password": newPassword}
	body, _ := json.Marshal(passwordData)

	req, _ := http.NewRequest("POST", server.URL+"/api/admin/password", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Password change request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		t.Fatalf("Expected status 200 on password change, got %d. Body: %s", res.StatusCode, string(bodyBytes))
	}

	var response APIResponse
	json.NewDecoder(res.Body).Decode(&response)
	if !response.Success || response.Data == nil {
		t.Fatal("Expected successful response with status data")
	}
	statusMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	if status, ok := statusMap["status"].(string); !ok || status != "Password changed" {
		t.Errorf("Expected status 'Password changed', got '%s'", statusMap["status"])
	}

	// Test that new password works for login
	loginDetails := map[string]string{"password": newPassword}
	loginBody, _ := json.Marshal(loginDetails)
	res, err = http.Post(server.URL+"/api/admin/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with new password, got %d", res.StatusCode)
	}

	// Test that old password no longer works
	oldLoginDetails := map[string]string{"password": "testpass"}
	oldLoginBody, _ := json.Marshal(oldLoginDetails)
	res, err = http.Post(server.URL+"/api/admin/login", "application/json", bytes.NewBuffer(oldLoginBody))
	if err != nil {
		t.Fatalf("Login with old password request failed: %v", err)
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with old password, got %d", res.StatusCode)
	}
}

func TestAPI_AdminPasswordChange_EmptyPassword(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)
	authHeader := "Bearer " + token

	// Test setting empty password (should be rejected by API)
	passwordData := map[string]string{"password": ""}
	body, _ := json.Marshal(passwordData)

	req, _ := http.NewRequest("POST", server.URL+"/api/admin/password", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Empty password change request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(res.Body)
		t.Fatalf("Expected status 400 for empty password change, got %d. Body: %s", res.StatusCode, string(bodyBytes))
	}

	// Verify the error response
	var errorRes APIResponse
	json.NewDecoder(res.Body).Decode(&errorRes)
	if errorRes.Success {
		t.Error("Expected error response for empty password")
	}
	if errorRes.Error == nil {
		t.Error("Expected error details in response")
	}
}

func TestAPI_AdminPasswordChange_PreservesOtherData(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	token := getAuthToken(t, server.URL)
	authHeader := "Bearer " + token

	// Add a link before changing password
	newLink := Link{ID: "preservetest", Title: "Preserve Test", URL: "https://preserve.com"}
	body, _ := json.Marshal(newLink)
	req, _ := http.NewRequest("POST", server.URL+"/api/admin/links", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Failed to add test link: %d", res.StatusCode)
	}

	// Change password
	passwordData := map[string]string{"password": "preservetestpass"}
	passwordBody, _ := json.Marshal(passwordData)
	req, _ = http.NewRequest("POST", server.URL+"/api/admin/password", bytes.NewBuffer(passwordBody))
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	res, _ = http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Failed to change password: %d", res.StatusCode)
	}

	// Verify the link still exists
	res, _ = http.Get(server.URL + "/api/links")
	var linksRes APIResponse
	json.NewDecoder(res.Body).Decode(&linksRes)
	if !linksRes.Success || linksRes.Data == nil {
		t.Fatal("Expected successful response with links data")
	}
	linksData, ok := linksRes.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	links, ok := linksData["links"].([]interface{})
	if !ok || len(links) != 2 {
		t.Errorf("Expected 2 links after password change, got %d", len(links))
	}

	// Verify the specific link exists
	found := false
	for _, linkInterface := range links {
		linkMap, ok := linkInterface.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := linkMap["id"].(string); ok && id == "preservetest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected test link to be preserved after password change")
	}
}

func TestAPI_RefreshConfig(t *testing.T) {
	server, teardown := setupTestServer(t)
	defer teardown()

	// Test public refresh-config endpoint (should work from localhost in tests)
	res, err := http.Post(server.URL+"/api/refresh-config", "application/json", nil)
	if err != nil {
		t.Fatalf("Refresh config request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		t.Fatalf("Expected status 200 on refresh config, got %d. Body: %s", res.StatusCode, string(bodyBytes))
	}

	var response APIResponse
	json.NewDecoder(res.Body).Decode(&response)
	if !response.Success || response.Data == nil {
		t.Fatal("Expected successful response with status data")
	}
	statusMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Invalid response data format")
	}
	if status, ok := statusMap["status"].(string); !ok || status != "Config refreshed from disk" {
		t.Errorf("Expected status 'Config refreshed from disk', got '%s'", statusMap["status"])
	}

	// Test wrong method (GET instead of POST)
	res, err = http.Get(server.URL + "/api/refresh-config")
	if err != nil {
		t.Fatalf("GET refresh config request failed: %v", err)
	}
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for wrong method, got %d", res.StatusCode)
	}
}

func TestAPI_RefreshConfig_LocalhostOnly(t *testing.T) {
	// This test would need a more complex setup to test non-localhost access
	// For now, we'll just test that the function exists and works correctly
	if !isLocalhost("127.0.0.1") {
		t.Error("Expected 127.0.0.1 to be recognized as localhost")
	}
	if !isLocalhost("::1") {
		t.Error("Expected ::1 to be recognized as localhost")
	}
	if !isLocalhost("localhost") {
		t.Error("Expected localhost to be recognized as localhost")
	}
	if !isLocalhost("127.0.0.1:8080") {
		t.Error("Expected 127.0.0.1:8080 to be recognized as localhost")
	}
	if isLocalhost("192.168.1.1") {
		t.Error("Expected 192.168.1.1 to NOT be recognized as localhost")
	}
	if isLocalhost("8.8.8.8") {
		t.Error("Expected 8.8.8.8 to NOT be recognized as localhost")
	}
}
