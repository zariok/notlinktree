package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	if id1 == "" || id2 == "" {
		t.Error("generateID() returned empty string")
	}
	if id1 == id2 {
		t.Error("generateID() should return unique values")
	}
}

func TestGenerateAdminToken(t *testing.T) {
	token1 := generateAdminToken()
	token2 := generateAdminToken()
	if token1 == "" || token2 == "" {
		t.Error("generateAdminToken() returned empty string")
	}
	if token1 == token2 {
		t.Error("generateAdminToken() should return unique values")
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Create a temp file
	tmpfile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Create a valid config object programmatically
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	initialConfig := Config{
		Admin: struct {
			Password string `yaml:"password" json:"-"`
		}{Password: string(hashedPassword)},
		Links: map[string]Link{
			"328bb7e6da77c3c2": {
				ID:          "328bb7e6da77c3c2",
				Title:       "Example Link",
				URL:         "https://example.com",
				Description: "An example link for testing",
				Type:        "Website",
				Clicks:      10,
			},
		},
		UI: UIConfig{
			Username:        "tester",
			Title:           "Test Links",
			PrimaryColor:    "#000000",
			SecondaryColor:  "#FFFFFF",
			BackgroundColor: "#EEEEEE",
		},
	}

	// Save it to the temp file
	if err := saveConfig(initialConfig, tmpfile.Name()); err != nil {
		t.Fatalf("Failed to write temp config: %v", err)
	}

	// Load config
	loaded, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}
	// Check that password is a valid bcrypt hash and matches
	if err := bcrypt.CompareHashAndPassword([]byte(loaded.Admin.Password), []byte("testpass")); err != nil {
		t.Errorf("Expected password 'testpass' to be hashed correctly, but got error: %v", err)
	}
	// Check links
	if len(loaded.Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(loaded.Links))
	}
	if _, ok := loaded.Links["328bb7e6da77c3c2"]; !ok {
		t.Errorf("Link with ID '328bb7e6da77c3c2' not loaded correctly")
	}
	if loaded.UI.Username != "tester" {
		t.Errorf("UI.Username not loaded correctly")
	}

	// Save config again to test the save function with a loaded config
	err = saveConfig(loaded, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	missingPath := "test/data/nonexistent_config.yaml"
	// Ensure the file does not exist before the test
	os.Remove(missingPath)
	_, err := loadConfig(missingPath)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist or nil, got %v", err)
	}
	// Clean up if the test created the file
	os.Remove(missingPath)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Use test/data/config_invalid.yaml
	_, err := loadConfig("test/data/config_invalid.yaml")
	t.Logf("YAML error: %v", err)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestAdminPasswordEdgeCases(t *testing.T) {
	cfg := Config{}
	cfg.Admin.Password = ""
	if cfg.Admin.Password != "" {
		t.Error("Expected empty password to be allowed")
	}
	long := make([]byte, 1024)
	for i := range long {
		long[i] = 'a'
	}
	cfg.Admin.Password = string(long)
	if len(cfg.Admin.Password) != 1024 {
		t.Error("Expected long password to be allowed")
	}
	cfg.Admin.Password = "!@#$%^&*()_+-=[]{}|;':,.<>/?"
	if cfg.Admin.Password == "" {
		t.Error("Expected special chars in password to be allowed")
	}
}

func BenchmarkSaveAndLoadConfig(b *testing.B) {
	tmpfile, err := os.CreateTemp("", "config_bench_*.yaml")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	hash, err := bcrypt.GenerateFromPassword([]byte("benchpass"), bcrypt.DefaultCost)
	if err != nil {
		b.Fatalf("Failed to hash password: %v", err)
	}

	cfg := Config{}
	cfg.Admin.Password = string(hash)
	cfg.Links = make(map[string]Link)
	cfg.Links["1"] = Link{ID: "1", Title: "Bench", URL: "https://bench", Description: "desc", Type: "Other", Clicks: 0}
	cfg.UI.Username = "bench"
	for i := 0; i < 10; i++ {
		id := generateID()
		cfg.Links[id] = Link{ID: id, Title: "L", URL: "https://l", Description: "d", Type: "Other", Clicks: 0}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := saveConfig(cfg, tmpfile.Name())
		if err != nil {
			b.Fatal(err)
		}
		_, err = loadConfig(tmpfile.Name())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_roundtrip_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	hash, _ := bcrypt.GenerateFromPassword([]byte("roundtrip"), bcrypt.DefaultCost)
	cfg := Config{}
	cfg.Admin.Password = string(hash)
	cfg.Links = make(map[string]Link)
	cfg.Links["1"] = Link{ID: "1", Title: "RT", URL: "https://rt", Description: "desc", Type: "Other", Clicks: 0}
	cfg.UI.Username = "rtuser"

	// Save config
	err = saveConfig(cfg, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveConfig failed: %v", err)
	}
	// Load config
	loaded, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}
	// Save loaded config
	err = saveConfig(loaded, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveConfig(loaded) failed: %v", err)
	}
	b1, _ := os.ReadFile(tmpfile.Name())
	b2, _ := os.ReadFile(tmpfile.Name())
	if !bytes.Equal(b1, b2) {
		t.Errorf("YAML round-trip failed: output mismatch")
	}
	// Also check struct equality for fields that should not be generated
	if cfg.Admin.Password != loaded.Admin.Password {
		t.Errorf("Password changed after round-trip")
	}
	if !reflect.DeepEqual(cfg.Links, loaded.Links) {
		t.Errorf("Links changed after round-trip")
	}
	if !reflect.DeepEqual(cfg.UI, loaded.UI) {
		t.Errorf("UI changed after round-trip")
	}
}

func TestLoadConfig_GeneratesPasswordIfMissing(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_missing_fields_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	// Write minimal config with no admin fields
	os.WriteFile(tmpfile.Name(), []byte("links: {}\nui: {}\n"), 0644)
	cfg, err := loadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}
	if cfg.Admin.Password == "" {
		t.Error("Expected password to be generated if missing")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cfg.Admin.Password), []byte("")); err == nil {
		t.Error("Generated password should not be an empty string's hash")
	}
}

func TestSaveAdminPassword(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_save_admin_pw_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Test with non-empty password
	testPassword := "testpassword123"
	err = saveAdminPassword(testPassword, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveAdminPassword failed: %v", err)
	}

	// Verify the password was saved correctly
	config, err := loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config after saving password: %v", err)
	}

	if config.Admin.Password == "" {
		t.Error("Expected password to be saved")
	}

	// Verify the password hash is correct
	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(testPassword))
	if err != nil {
		t.Errorf("Saved password hash does not match original password: %v", err)
	}

	// Test with empty password
	err = saveAdminPassword("", tmpfile.Name())
	if err != nil {
		t.Fatalf("saveAdminPassword with empty password failed: %v", err)
	}

	// Verify empty password was saved
	config, err = loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config after saving empty password: %v", err)
	}

	if config.Admin.Password == "" {
		t.Error("Expected empty password to be saved as hash")
	}

	// Verify the empty password hash is correct
	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(""))
	if err != nil {
		t.Errorf("Saved empty password hash does not match: %v", err)
	}
}

func TestSaveAdminPassword_WithExistingConfig(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_save_admin_pw_existing_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Create an existing config with some data
	existingConfig := Config{
		Links: map[string]Link{
			"testlink": {
				ID:          "testlink",
				Title:       "Test Link",
				URL:         "https://example.com",
				Description: "A test link",
				Type:        "Website",
				Clicks:      5,
			},
		},
		UI: UIConfig{
			Username: "testuser",
			Title:    "Test Title",
		},
	}

	// Save the existing config
	err = saveConfig(existingConfig, tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to save existing config: %v", err)
	}

	// Now save a new admin password
	newPassword := "newadminpass456"
	err = saveAdminPassword(newPassword, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveAdminPassword failed: %v", err)
	}

	// Verify the password was updated but other data was preserved
	config, err := loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config after password update: %v", err)
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(newPassword))
	if err != nil {
		t.Errorf("Updated password hash does not match: %v", err)
	}

	// Check that existing data was preserved
	if len(config.Links) != 1 {
		t.Errorf("Expected 1 link to be preserved, got %d", len(config.Links))
	}
	if config.Links["testlink"].Title != "Test Link" {
		t.Errorf("Expected link title to be preserved, got %s", config.Links["testlink"].Title)
	}
	if config.UI.Username != "testuser" {
		t.Errorf("Expected UI username to be preserved, got %s", config.UI.Username)
	}
}

func TestLoadConfigForPasswordSet(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_load_for_pw_set_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Test loading non-existent file (should create empty config)
	config, err := loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfigForPasswordSet failed for non-existent file: %v", err)
	}
	if config.Admin.Password != "" {
		t.Error("Expected empty password for new config")
	}
	if config.Links == nil {
		t.Error("Expected empty links map for new config")
	}

	// Test loading existing file
	testConfig := Config{
		Admin: struct {
			Password string `yaml:"password" json:"-"`
		}{
			Password: "hashedpassword",
		},
		Links: map[string]Link{
			"test": {ID: "test", Title: "Test", URL: "https://test.com"},
		},
		UI: UIConfig{Username: "testuser"},
	}

	err = saveConfig(testConfig, tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	loadedConfig, err := loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("loadConfigForPasswordSet failed for existing file: %v", err)
	}

	if loadedConfig.Admin.Password != "hashedpassword" {
		t.Errorf("Expected password 'hashedpassword', got '%s'", loadedConfig.Admin.Password)
	}
	if len(loadedConfig.Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(loadedConfig.Links))
	}
	if loadedConfig.UI.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", loadedConfig.UI.Username)
	}
}

func TestLoadConfigForPasswordSet_InvalidYAML(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_invalid_yaml_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write invalid YAML
	os.WriteFile(tmpfile.Name(), []byte("invalid: yaml: content: [\n"), 0644)

	_, err = loadConfigForPasswordSet(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestSaveAdminPassword_EdgeCases(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_save_admin_pw_edge_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Test with long password (but within bcrypt's 72-byte limit)
	longPassword := string(make([]byte, 70))
	for i := range longPassword {
		longPassword = longPassword[:i] + "a" + longPassword[i+1:]
	}

	err = saveAdminPassword(longPassword, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveAdminPassword failed with long password: %v", err)
	}

	// Verify it was saved correctly
	config, err := loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config with long password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(longPassword))
	if err != nil {
		t.Errorf("Long password hash does not match: %v", err)
	}

	// Test with special characters
	specialPassword := "!@#$%^&*()_+-=[]{}|;':\",./<>?`~"
	err = saveAdminPassword(specialPassword, tmpfile.Name())
	if err != nil {
		t.Fatalf("saveAdminPassword failed with special characters: %v", err)
	}

	config, err = loadConfigForPasswordSet(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config with special characters: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(specialPassword))
	if err != nil {
		t.Errorf("Special character password hash does not match: %v", err)
	}

	// Test with password that exceeds bcrypt's 72-byte limit
	tooLongPassword := string(make([]byte, 100))
	for i := range tooLongPassword {
		tooLongPassword = tooLongPassword[:i] + "a" + tooLongPassword[i+1:]
	}

	err = saveAdminPassword(tooLongPassword, tmpfile.Name())
	if err == nil {
		t.Error("Expected error for password exceeding bcrypt's 72-byte limit, got nil")
	}
}
