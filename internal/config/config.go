package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// API Keys
	FinnhubAPIKey string
	EDGARUserAgent string
	JWTSecret     string
	
	// Feature Flags
	UseMockData bool
	
	// API Settings
	RequestTimeout int // seconds
}

// Global config instance
var AppConfig *Config

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		FinnhubAPIKey:  os.Getenv("FINNHUB_API_KEY"),
		EDGARUserAgent: os.Getenv("EDGAR_USER_AGENT"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		UseMockData:    strings.ToLower(os.Getenv("USE_MOCK_DATA")) == "true",
		RequestTimeout: 10, // default 10 seconds
	}
	
	// Validate required configuration
	var missingVars []string
	
	if config.JWTSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}
	
	// EDGAR User-Agent is required by SEC (they block requests without it)
	if config.EDGARUserAgent == "" {
		// Provide a helpful default if not set
		config.EDGARUserAgent = "finEdSkywalker/1.0"
	}
	
	// Finnhub API key is optional if using mock data
	if config.FinnhubAPIKey == "" && !config.UseMockData {
		missingVars = append(missingVars, "FINNHUB_API_KEY (or set USE_MOCK_DATA=true)")
	}
	
	if len(missingVars) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}
	
	AppConfig = config
	return config, nil
}

// GetConfig returns the global config instance
func GetConfig() *Config {
	if AppConfig == nil {
		// Try to load if not already loaded
		config, err := Load()
		if err != nil {
			// Return a config with defaults for graceful degradation
			return &Config{
				UseMockData:    true,
				RequestTimeout: 10,
				EDGARUserAgent: "finEdSkywalker/1.0",
			}
		}
		return config
	}
	return AppConfig
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	
	if c.FinnhubAPIKey == "" && !c.UseMockData {
		return fmt.Errorf("FINNHUB_API_KEY is required (or enable mock data mode)")
	}
	
	return nil
}

// IsMockMode returns true if the application is running in mock data mode
func (c *Config) IsMockMode() bool {
	return c.UseMockData
}

