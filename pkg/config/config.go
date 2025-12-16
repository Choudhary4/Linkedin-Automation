package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	// LinkedIn credentials
	LinkedIn LinkedInConfig `yaml:"linkedin"`

	// Application settings
	App AppConfig `yaml:"app"`

	// Browser settings
	Browser BrowserConfig `yaml:"browser"`

	// Connection request settings
	Connections ConnectionConfig `yaml:"connections"`

	// Messaging settings
	Messaging MessagingConfig `yaml:"messaging"`

	// Search configuration
	Search SearchConfig `yaml:"search"`

	// Stealth configuration
	Stealth StealthConfig `yaml:"stealth"`

	// Database configuration
	Database DatabaseConfig `yaml:"database"`

	// Session configuration
	Session SessionConfig `yaml:"session"`
}

type LinkedInConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type AppConfig struct {
	LogLevel string `yaml:"log_level"`
}

type BrowserConfig struct {
	Headless   bool `yaml:"headless"`
	SlowMotion int  `yaml:"slow_motion"`
}

type ConnectionConfig struct {
	MaxPerDay  int `yaml:"max_per_day"`
	MaxPerHour int `yaml:"max_per_hour"`
}

type MessagingConfig struct {
	MaxPerDay              int    `yaml:"max_per_day"`
	DefaultMessageTemplate string `yaml:"default_message_template"`
}

type SearchConfig struct {
	DefaultKeywords []string `yaml:"default_keywords"`
	DefaultLocation string   `yaml:"default_location"`
}

type StealthConfig struct {
	MinActionDelay         int  `yaml:"min_action_delay"`
	MaxActionDelay         int  `yaml:"max_action_delay"`
	EnableRandomScroll     bool `yaml:"enable_random_scroll"`
	EnableTypingSimulation bool `yaml:"enable_typing_simulation"`
	EnableMouseHover       bool `yaml:"enable_mouse_hover"`
	OperatingHoursStart    int  `yaml:"operating_hours_start"`
	OperatingHoursEnd      int  `yaml:"operating_hours_end"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type SessionConfig struct {
	Path string `yaml:"path"`
}

// LoadConfig loads configuration from environment variables and YAML file
func LoadConfig(yamlPath string) (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{}

	// Try to load from YAML file if provided
	if yamlPath != "" {
		if err := loadFromYAML(yamlPath, config); err != nil {
			return nil, fmt.Errorf("failed to load YAML config: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// loadFromYAML loads configuration from a YAML file
func loadFromYAML(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// LinkedIn credentials
	if email := os.Getenv("LINKEDIN_EMAIL"); email != "" {
		config.LinkedIn.Email = email
	}
	if password := os.Getenv("LINKEDIN_PASSWORD"); password != "" {
		config.LinkedIn.Password = password
	}

	// App settings
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.App.LogLevel = logLevel
	} else if config.App.LogLevel == "" {
		config.App.LogLevel = "info"
	}

	// Browser settings
	if headless := os.Getenv("HEADLESS"); headless != "" {
		config.Browser.Headless = headless == "true"
	}
	if slowMotion := os.Getenv("SLOW_MOTION"); slowMotion != "" {
		if val, err := strconv.Atoi(slowMotion); err == nil {
			config.Browser.SlowMotion = val
		}
	}

	// Connection settings
	if maxPerDay := os.Getenv("MAX_CONNECTIONS_PER_DAY"); maxPerDay != "" {
		if val, err := strconv.Atoi(maxPerDay); err == nil {
			config.Connections.MaxPerDay = val
		}
	} else if config.Connections.MaxPerDay == 0 {
		config.Connections.MaxPerDay = 20
	}
	if maxPerHour := os.Getenv("MAX_CONNECTIONS_PER_HOUR"); maxPerHour != "" {
		if val, err := strconv.Atoi(maxPerHour); err == nil {
			config.Connections.MaxPerHour = val
		}
	} else if config.Connections.MaxPerHour == 0 {
		config.Connections.MaxPerHour = 5
	}

	// Messaging settings
	if maxPerDay := os.Getenv("MAX_MESSAGES_PER_DAY"); maxPerDay != "" {
		if val, err := strconv.Atoi(maxPerDay); err == nil {
			config.Messaging.MaxPerDay = val
		}
	} else if config.Messaging.MaxPerDay == 0 {
		config.Messaging.MaxPerDay = 15
	}
	if template := os.Getenv("DEFAULT_MESSAGE_TEMPLATE"); template != "" {
		config.Messaging.DefaultMessageTemplate = template
	} else if config.Messaging.DefaultMessageTemplate == "" {
		config.Messaging.DefaultMessageTemplate = "Hi {firstName}, I'd love to connect with you!"
	}

	// Search settings
	if keywords := os.Getenv("DEFAULT_SEARCH_KEYWORDS"); keywords != "" {
		config.Search.DefaultKeywords = strings.Split(keywords, ",")
		for i := range config.Search.DefaultKeywords {
			config.Search.DefaultKeywords[i] = strings.TrimSpace(config.Search.DefaultKeywords[i])
		}
	}
	if location := os.Getenv("DEFAULT_LOCATION"); location != "" {
		config.Search.DefaultLocation = location
	}

	// Stealth settings
	if minDelay := os.Getenv("MIN_ACTION_DELAY"); minDelay != "" {
		if val, err := strconv.Atoi(minDelay); err == nil {
			config.Stealth.MinActionDelay = val
		}
	} else if config.Stealth.MinActionDelay == 0 {
		config.Stealth.MinActionDelay = 2000
	}
	if maxDelay := os.Getenv("MAX_ACTION_DELAY"); maxDelay != "" {
		if val, err := strconv.Atoi(maxDelay); err == nil {
			config.Stealth.MaxActionDelay = val
		}
	} else if config.Stealth.MaxActionDelay == 0 {
		config.Stealth.MaxActionDelay = 5000
	}
	if enableScroll := os.Getenv("ENABLE_RANDOM_SCROLL"); enableScroll != "" {
		config.Stealth.EnableRandomScroll = enableScroll == "true"
	} else {
		config.Stealth.EnableRandomScroll = true
	}
	if enableTyping := os.Getenv("ENABLE_TYPING_SIMULATION"); enableTyping != "" {
		config.Stealth.EnableTypingSimulation = enableTyping == "true"
	} else {
		config.Stealth.EnableTypingSimulation = true
	}
	if enableHover := os.Getenv("ENABLE_MOUSE_HOVER"); enableHover != "" {
		config.Stealth.EnableMouseHover = enableHover == "true"
	} else {
		config.Stealth.EnableMouseHover = true
	}
	if startHour := os.Getenv("OPERATING_HOURS_START"); startHour != "" {
		if val, err := strconv.Atoi(startHour); err == nil {
			config.Stealth.OperatingHoursStart = val
		}
	} else if config.Stealth.OperatingHoursStart == 0 {
		config.Stealth.OperatingHoursStart = 0 // Default: 24h operation (no time restriction)
	}
	if endHour := os.Getenv("OPERATING_HOURS_END"); endHour != "" {
		if val, err := strconv.Atoi(endHour); err == nil {
			config.Stealth.OperatingHoursEnd = val
		}
	} else if config.Stealth.OperatingHoursEnd == 0 {
		config.Stealth.OperatingHoursEnd = 24 // Default: 24h operation (no time restriction)
	}

	// Database settings
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		config.Database.Path = dbPath
	} else if config.Database.Path == "" {
		config.Database.Path = "./data/subspace.db"
	}

	// Session settings
	if sessionPath := os.Getenv("SESSION_PATH"); sessionPath != "" {
		config.Session.Path = sessionPath
	} else if config.Session.Path == "" {
		config.Session.Path = "./data/sessions"
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.LinkedIn.Email == "" {
		return fmt.Errorf("LinkedIn email is required")
	}
	if c.LinkedIn.Password == "" {
		return fmt.Errorf("LinkedIn password is required")
	}
	if c.Connections.MaxPerDay <= 0 {
		return fmt.Errorf("max connections per day must be positive")
	}
	if c.Connections.MaxPerHour <= 0 {
		return fmt.Errorf("max connections per hour must be positive")
	}
	if c.Stealth.MinActionDelay < 0 {
		return fmt.Errorf("min action delay cannot be negative")
	}
	if c.Stealth.MaxActionDelay < c.Stealth.MinActionDelay {
		return fmt.Errorf("max action delay must be greater than or equal to min action delay")
	}
	if c.Stealth.OperatingHoursStart < 0 || c.Stealth.OperatingHoursStart > 23 {
		return fmt.Errorf("operating hours start must be between 0 and 23")
	}
	if c.Stealth.OperatingHoursEnd < 0 || c.Stealth.OperatingHoursEnd > 23 {
		return fmt.Errorf("operating hours end must be between 0 and 23")
	}
	return nil
}
