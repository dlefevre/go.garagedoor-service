package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	// All known configuration properties, and weither they are mandatory or not
	KnownKeys = map[string]bool{
		"mode":            true,
		"bind.port":       true,
		"bind.host":       true,
		"gpio.toggle_pin": true,
		"gpio.open_pin":   true,
		"gpio.closed_pin": true,
		"api_keys":        true,
	}

	viperInst *viper.Viper = nil
	once      sync.Once
)

// Create a new Viper instance and load the configuration file.
func loadConfig() {
	viperInst = viper.New()

	viperInst.SetConfigName("config")
	viperInst.SetConfigType("yaml")
	configPath := os.Getenv("GARAGESERVICE_CONFIG_PATH")
	if configPath != "" {
		viperInst.AddConfigPath(configPath)
	}
	viperInst.AddConfigPath(".")

	if err := viperInst.ReadInConfig(); err != nil {
		panic(fmt.Errorf("config: fatal error while parsing config file: %s", err))
	}
}

// Verifies that all mandatory keys are set in the configuration file,
// and that no unknown keys are present.
func verifyKeys() error {
	for key, mandatory := range KnownKeys {
		if mandatory && !viperInst.IsSet(key) {
			return fmt.Errorf("config: configuration property %s is mandatory", key)
		}
	}
	for _, key := range viperInst.AllKeys() {
		if _, found := KnownKeys[key]; !found {
			return fmt.Errorf("config: configuration property %s is unknown", key)
		}
	}
	return nil
}

// Verify that the configuration is valid.
func Verify() error {
	once.Do(loadConfig)
	if err := verifyKeys(); err != nil {
		return err
	}
	mode := viperInst.GetString("mode")
	if mode != "development" && mode != "production" {
		return fmt.Errorf("config: mode must be either 'development' or 'production'")
	}
	port := viperInst.GetInt("bind.port")
	if port < 0 || port > 65535 {
		return fmt.Errorf("config: bind.port must be a valid port number")
	}
	togglePin := viperInst.GetInt("gpio.toggle_pin")
	if togglePin < 0 {
		return fmt.Errorf("config: gpio.toggle_pin must be a valid pin number")
	}
	openPin := viperInst.GetInt("gpio.open_pin")
	if openPin < 0 {
		return fmt.Errorf("config: gpio.open_pin must be a valid pin number")
	}
	closedPin := viperInst.GetInt("gpio.closed_pin")
	if closedPin < 0 {
		return fmt.Errorf("config: gpio.closed_pin must be a valid pin number")
	}
	apiKeys := viperInst.GetStringSlice("api_keys")
	if len(apiKeys) == 0 {
		return fmt.Errorf("config: api_keys must contain at least one key")
	}

	return nil
}

// GetMode returns the current mode.
func GetMode() string {
	once.Do(loadConfig)
	return viperInst.GetString("mode")
}

// GetBindPort returns the port to bind the web server to.
func GetBindPort() int {
	once.Do(loadConfig)
	return viperInst.GetInt("bind.port")
}

// GetBindHost returns the host to bind the web server to.
func GetBindHost() string {
	once.Do(loadConfig)
	return viperInst.GetString("bind.host")
}

// GetTogglePin returns the GPIO pin number for the toggle state.
func GetTogglePin() int {
	once.Do(loadConfig)
	return viperInst.GetInt("gpio.toggle_pin")
}

// GetOpenPin returns the GPIO pin number for the open state.
func GetOpenPin() int {
	once.Do(loadConfig)
	return viperInst.GetInt("gpio.open_pin")
}

// GetClosedPin returns the GPIO pin number for the closed state.
func GetClosedPin() int {
	once.Do(loadConfig)
	return viperInst.GetInt("gpio.closed_pin")
}

// GetApiKeys returns the list of API keys.
func GetApiKeys() []string {
	once.Do(loadConfig)
	return viperInst.GetStringSlice("api_keys")
}
