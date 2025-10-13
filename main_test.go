package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestSetAdminPWFlag(t *testing.T) {
	// Test that the flag is properly defined
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	setAdminPW := fs.String("setadminpw", "", "Set the admin password and exit")

	// Test parsing with the flag
	err := fs.Parse([]string{"-setadminpw", "testpassword123"})
	if err != nil {
		t.Fatalf("Failed to parse -setadminpw flag: %v", err)
	}

	if *setAdminPW != "testpassword123" {
		t.Errorf("Expected flag value 'testpassword123', got '%s'", *setAdminPW)
	}
}

func TestSetAdminPWFlag_EmptyValue(t *testing.T) {
	// Test that the flag works with empty values
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	setAdminPW := fs.String("setadminpw", "", "Set the admin password and exit")

	// Test parsing with empty value
	err := fs.Parse([]string{"-setadminpw", ""})
	if err != nil {
		t.Fatalf("Failed to parse -setadminpw flag with empty value: %v", err)
	}

	if *setAdminPW != "" {
		t.Errorf("Expected empty flag value, got '%s'", *setAdminPW)
	}
}

func TestSetAdminPWFlag_NotProvided(t *testing.T) {
	// Test that the flag has default value when not provided
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	setAdminPW := fs.String("setadminpw", "", "Set the admin password and exit")

	// Test parsing without the flag
	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Failed to parse without -setadminpw flag: %v", err)
	}

	if *setAdminPW != "" {
		t.Errorf("Expected default empty flag value, got '%s'", *setAdminPW)
	}
}

func TestFlagVisit_DetectsSetAdminPW(t *testing.T) {
	// Test that flag.Visit correctly detects when -setadminpw is provided
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	setAdminPW := fs.String("setadminpw", "", "Set the admin password and exit")

	// Parse with the flag
	err := fs.Parse([]string{"-setadminpw", "testpass"})
	if err != nil {
		t.Fatalf("Failed to parse -setadminpw flag: %v", err)
	}

	// Check if flag was visited
	setAdminPWProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "setadminpw" {
			setAdminPWProvided = true
		}
	})

	if !setAdminPWProvided {
		t.Error("Expected -setadminpw flag to be detected as provided")
	}

	if *setAdminPW != "testpass" {
		t.Errorf("Expected flag value 'testpass', got '%s'", *setAdminPW)
	}
}

func TestFlagVisit_DetectsSetAdminPW_EmptyValue(t *testing.T) {
	// Test that flag.Visit correctly detects when -setadminpw is provided even with empty value
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	setAdminPW := fs.String("setadminpw", "", "Set the admin password and exit")

	// Parse with empty value
	err := fs.Parse([]string{"-setadminpw", ""})
	if err != nil {
		t.Fatalf("Failed to parse -setadminpw flag with empty value: %v", err)
	}

	// Check if flag was visited
	setAdminPWProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "setadminpw" {
			setAdminPWProvided = true
		}
	})

	if !setAdminPWProvided {
		t.Error("Expected -setadminpw flag to be detected as provided even with empty value")
	}

	if *setAdminPW != "" {
		t.Errorf("Expected empty flag value, got '%s'", *setAdminPW)
	}
}

func TestFlagVisit_NotProvided(t *testing.T) {
	// Test that flag.Visit does not detect -setadminpw when not provided
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	setAdminPW := fs.String("setadminpw", "", "Set the admin password and exit")

	// Parse without the flag
	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Failed to parse without -setadminpw flag: %v", err)
	}

	// Check if flag was visited
	setAdminPWProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "setadminpw" {
			setAdminPWProvided = true
		}
	})

	if setAdminPWProvided {
		t.Error("Expected -setadminpw flag to NOT be detected as provided when not provided")
	}

	if *setAdminPW != "" {
		t.Errorf("Expected default empty flag value, got '%s'", *setAdminPW)
	}
}

func TestSetAdminPWIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "notlinktree-integration-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up environment
	originalDataDir := os.Getenv("NLT_DATA")
	os.Setenv("NLT_DATA", tempDir)
	defer func() {
		if originalDataDir == "" {
			os.Unsetenv("NLT_DATA")
		} else {
			os.Setenv("NLT_DATA", originalDataDir)
		}
	}()

	configPath := filepath.Join(tempDir, "config.yaml")

	// Test setting a password
	testPassword := "integrationtest123"
	err = saveAdminPassword(testPassword, configPath)
	if err != nil {
		t.Fatalf("saveAdminPassword failed: %v", err)
	}

	// Verify the password was saved
	config, err := loadConfigForPasswordSet(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Admin.Password == "" {
		t.Error("Expected password to be saved")
	}

	// Verify the password hash is correct
	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(testPassword))
	if err != nil {
		t.Errorf("Saved password hash does not match: %v", err)
	}

	// Test updating the password
	newPassword := "updatedpassword456"
	err = saveAdminPassword(newPassword, configPath)
	if err != nil {
		t.Fatalf("saveAdminPassword update failed: %v", err)
	}

	// Verify the password was updated
	config, err = loadConfigForPasswordSet(configPath)
	if err != nil {
		t.Fatalf("Failed to load updated config: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(newPassword))
	if err != nil {
		t.Errorf("Updated password hash does not match: %v", err)
	}

	// Verify old password no longer works
	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(testPassword))
	if err == nil {
		t.Error("Old password should no longer work")
	}
}

func TestSetAdminPWIntegration_WithExistingConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "notlinktree-integration-existing-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up environment
	originalDataDir := os.Getenv("NLT_DATA")
	os.Setenv("NLT_DATA", tempDir)
	defer func() {
		if originalDataDir == "" {
			os.Unsetenv("NLT_DATA")
		} else {
			os.Setenv("NLT_DATA", originalDataDir)
		}
	}()

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create an existing config with data
	existingConfig := Config{
		Links: map[string]Link{
			"existinglink": {
				ID:          "existinglink",
				Title:       "Existing Link",
				URL:         "https://existing.com",
				Description: "An existing link",
				Type:        "Website",
				Clicks:      10,
			},
		},
		UI: UIConfig{
			Username:        "existinguser",
			Title:           "Existing Title",
			PrimaryColor:    "#FF0000",
			SecondaryColor:  "#00FF00",
			BackgroundColor: "#0000FF",
		},
	}

	// Save the existing config
	err = saveConfig(existingConfig, configPath)
	if err != nil {
		t.Fatalf("Failed to save existing config: %v", err)
	}

	// Now set a new admin password
	newPassword := "newadminpass789"
	err = saveAdminPassword(newPassword, configPath)
	if err != nil {
		t.Fatalf("saveAdminPassword failed: %v", err)
	}

	// Verify the password was set
	config, err := loadConfigForPasswordSet(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(newPassword))
	if err != nil {
		t.Errorf("New password hash does not match: %v", err)
	}

	// Verify existing data was preserved
	if len(config.Links) != 1 {
		t.Errorf("Expected 1 existing link, got %d", len(config.Links))
	}
	if config.Links["existinglink"].Title != "Existing Link" {
		t.Errorf("Expected existing link title to be preserved, got '%s'", config.Links["existinglink"].Title)
	}
	if config.UI.Username != "existinguser" {
		t.Errorf("Expected existing username to be preserved, got '%s'", config.UI.Username)
	}
	if config.UI.PrimaryColor != "#FF0000" {
		t.Errorf("Expected existing primary color to be preserved, got '%s'", config.UI.PrimaryColor)
	}
}

func TestSetAdminPWIntegration_EmptyPassword(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "notlinktree-integration-empty-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up environment
	originalDataDir := os.Getenv("NLT_DATA")
	os.Setenv("NLT_DATA", tempDir)
	defer func() {
		if originalDataDir == "" {
			os.Unsetenv("NLT_DATA")
		} else {
			os.Setenv("NLT_DATA", originalDataDir)
		}
	}()

	configPath := filepath.Join(tempDir, "config.yaml")

	// Test setting an empty password
	err = saveAdminPassword("", configPath)
	if err != nil {
		t.Fatalf("saveAdminPassword with empty password failed: %v", err)
	}

	// Verify the empty password was saved
	config, err := loadConfigForPasswordSet(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.Admin.Password == "" {
		t.Error("Expected empty password to be saved as hash")
	}

	// Verify the empty password hash is correct
	err = bcrypt.CompareHashAndPassword([]byte(config.Admin.Password), []byte(""))
	if err != nil {
		t.Errorf("Empty password hash does not match: %v", err)
	}
}

func TestSetAdminPWIntegration_ErrorHandling(t *testing.T) {
	// Test with invalid path (directory that doesn't exist)
	invalidPath := "/nonexistent/directory/config.yaml"
	err := saveAdminPassword("testpass", invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}

	// Test with read-only directory (if possible on this system)
	tempDir, err := os.MkdirTemp("", "notlinktree-readonly-")
	if err != nil {
		t.Skip("Skipping read-only test: failed to create temp dir")
	}
	defer os.RemoveAll(tempDir)

	// Make directory read-only
	err = os.Chmod(tempDir, 0444)
	if err != nil {
		t.Skip("Skipping read-only test: failed to make directory read-only")
	}
	defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup

	readOnlyPath := filepath.Join(tempDir, "config.yaml")
	err = saveAdminPassword("testpass", readOnlyPath)
	if err == nil {
		t.Error("Expected error for read-only directory, got nil")
	}
}

func TestReloadRunningInstance(t *testing.T) {
	// Test with default port
	originalPort := os.Getenv("NLT_PORT")
	os.Unsetenv("NLT_PORT")
	defer func() {
		if originalPort == "" {
			os.Unsetenv("NLT_PORT")
		} else {
			os.Setenv("NLT_PORT", originalPort)
		}
	}()

	// Test with no server running (should fail)
	err := reloadRunningInstance()
	if err == nil {
		t.Error("Expected error when no server is running, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("Expected connection refused error, got: %v", err)
	}

	// Test with custom port
	os.Setenv("NLT_PORT", "9999")
	err = reloadRunningInstance()
	if err == nil {
		t.Error("Expected error when no server is running on custom port, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("Expected connection refused error, got: %v", err)
	}
}

func TestCLI_SetAdminPW_Validation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "notlinktree-cli-validation-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set up environment
	originalDataDir := os.Getenv("NLT_DATA")
	os.Setenv("NLT_DATA", tempDir)
	defer func() {
		if originalDataDir == "" {
			os.Unsetenv("NLT_DATA")
		} else {
			os.Setenv("NLT_DATA", originalDataDir)
		}
	}()

	configPath := filepath.Join(tempDir, "config.yaml")

	testCases := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
		{
			name:        "too short password",
			password:    "abc123",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "whitespace only password",
			password:    "   \t\n  ",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "previously blocked common password with numbers",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "no letters",
			password:    "12345678",
			expectError: true,
			errorMsg:    "password must contain at least one letter",
		},
		{
			name:        "no numbers",
			password:    "abcdefgh",
			expectError: true,
			errorMsg:    "password must contain at least one number",
		},
		{
			name:        "valid password",
			password:    "mypassword123",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the validation function directly
			err := validatePassword(tc.password)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil for password: %q", tc.password)
					return
				}
				if tc.errorMsg != "" && err.Error() != tc.errorMsg {
					t.Errorf("Expected error message %q, got %q", tc.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v for password: %q", err, tc.password)
				}
			}

			// Test the saveAdminPassword function with validation
			if !tc.expectError {
				// Only test successful cases with saveAdminPassword
				err = saveAdminPassword(tc.password, configPath)
				if err != nil {
					t.Errorf("saveAdminPassword failed for valid password: %v", err)
				}
			}
		})
	}
}
