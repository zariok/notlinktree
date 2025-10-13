package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

type UIConfig struct {
	Username        string `yaml:"username" json:"username"`
	Title           string `yaml:"title" json:"title"`
	PrimaryColor    string `yaml:"primaryColor" json:"primaryColor"`
	SecondaryColor  string `yaml:"secondaryColor" json:"secondaryColor"`
	BackgroundColor string `yaml:"backgroundColor" json:"backgroundColor"`
}

type Config struct {
	Admin struct {
		Password string `yaml:"password" json:"-"`
	} `yaml:"admin"`
	Links map[string]Link `yaml:"links"`
	UI    UIConfig        `yaml:"ui" json:"ui"`
}

type Link struct {
	ID          string `json:"id" yaml:"id"`
	Title       string `json:"title" yaml:"title"`
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description" yaml:"description"`
	Type        string `json:"type" yaml:"type"`
	Clicks      int    `json:"clicks" yaml:"clicks"`
}

func generateAdminToken() string {
	b := make([]byte, 9)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func loadConfig(configPath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if it doesn't exist
			config = Config{
				Links: make(map[string]Link),
			}
			// Generate new admin password
			pw := generateAdminToken()
			hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
			if err != nil {
				return Config{}, fmt.Errorf("could not hash password: %w", err)
			}
			config.Admin.Password = string(hash)
			log.Printf("New admin password has been generated and saved to %s.", configPath)
			log.Printf("One-time admin password: %s", pw)
			return config, saveConfig(config, configPath)
		}
		return Config{}, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	// If admin password is not set, generate a new one
	if config.Admin.Password == "" {
		pw := generateAdminToken()
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return Config{}, fmt.Errorf("could not hash password: %w", err)
		}
		config.Admin.Password = string(hash)
		log.Printf("Generated new admin password and saved to %s.", configPath)
		log.Printf("One-time admin password: %s", pw)
		err = saveConfig(config, configPath)
		if err != nil {
			return config, err
		}
		return loadConfig(configPath)
	}

	return config, nil
}

func saveConfig(config Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// saveAdminPassword hashes the provided password and saves it to the config
func saveAdminPassword(password string, configPath string) error {
	// Load existing config or create new one
	config, err := loadConfigForPasswordSet(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not hash password: %w", err)
	}

	// Update config with new password
	config.Admin.Password = string(hash)

	// Save the updated config
	return saveConfig(config, configPath)
}

// loadConfigForPasswordSet loads config without requiring JWT secret (for password setting)
func loadConfigForPasswordSet(configPath string) (Config, error) {
	var config Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config if it doesn't exist
			config = Config{
				Links: make(map[string]Link),
			}
			return config, nil
		}
		return Config{}, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	// Ensure Links is initialized even if not present in YAML
	if config.Links == nil {
		config.Links = make(map[string]Link)
	}

	return config, nil
}

// validatePassword checks if a password meets the minimum requirements
func validatePassword(password string) error {
	const minLength = 8

	// Check for empty password
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Check minimum length
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters long", minLength)
	}

	// Check for whitespace-only password
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password cannot be only whitespace")
	}

	// Check for at least one letter and one number
	hasLetter := false
	hasNumber := false

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
	}

	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter")
	}

	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	return nil
}
