package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	// All known configuration properties, and weither they are mandatory or not
	knownKeys = map[string]bool{
		"mode":              true,
		"bind.port":         true,
		"bind.host":         true,
		"gpio.toggle_pin":   true,
		"gpio.open_pin":     true,
		"gpio.closed_pin":   true,
		"api_keys":          true,
		"mqtt.enabled":      true,
		"mqtt.url":          false,
		"mqtt.username":     false,
		"mqtt.password":     false,
		"mqtt.client_id":    false,
		"mqtt.topic_prefix": false,
	}

	viperInst *viper.Viper
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
	for key, mandatory := range knownKeys {
		if mandatory && !viperInst.IsSet(key) {
			return fmt.Errorf("config: configuration property %s is mandatory", key)
		}
	}
	for _, key := range viperInst.AllKeys() {
		if _, found := knownKeys[key]; !found {
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
	mqttEnabled := viperInst.GetBool("mqtt.enabled")
	if mqttEnabled {
		mqttURL := viperInst.GetString("mqtt.url")
		if mqttURL == "" {
			return fmt.Errorf("config: mqtt.url must be set when mqtt.enabled is true")
		}
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

// GetAPIKeys returns the list of API keys.
func GetAPIKeys() []string {
	once.Do(loadConfig)
	return viperInst.GetStringSlice("api_keys")
}

// GetMQTTEnabled returns whether MQTT is enabled.
func GetMQTTEnabled() bool {
	once.Do(loadConfig)
	return viperInst.GetBool("mqtt.enabled")
}

// GetMQTTURL returns the URL for the MQTT broker.
func GetMQTTURL() string {
	once.Do(loadConfig)
	if !viperInst.IsSet("mqtt.url") {
		return ""
	}
	return viperInst.GetString("mqtt.url")
}

// GetMQTTUsername returns the client ID for the MQTT client.
func GetMQTTUsername() string {
	once.Do(loadConfig)
	if !viperInst.IsSet("mqtt.username") {
		return ""
	}
	return viperInst.GetString("mqtt.username")
}

// GetMQTTPassword returns the password for the MQTT client.
func GetMQTTPassword() string {
	once.Do(loadConfig)
	if !viperInst.IsSet("mqtt.password") {
		return ""
	}
	return viperInst.GetString("mqtt.password")
}

// GetMQTTClientID returns the client ID for the MQTT client.
func GetMQTTClientID() string {
	once.Do(loadConfig)
	if !viperInst.IsSet("mqtt.client_id") {
		return ""
	}
	return viperInst.GetString("mqtt.client_id")
}

func GetMQTTTopicPrefix() string {
	once.Do(loadConfig)
	if !viperInst.IsSet("mqtt.topic_prefix") {
		return ""
	}
	return viperInst.GetString("mqtt.topic_prefix")
}
