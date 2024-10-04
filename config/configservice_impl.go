package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

var (
	// All known configuration properties, and weither they are mandatory or not
	KNOWN_KEYS = map[string]bool{
		"mode":            true,
		"bind.port":       true,
		"bind.host":       true,
		"gpio.toggle_pin": true,
		"gpio.open_pin":   true,
		"gpio.closed_pin": true,
		"api_keys":        true,
	}

	lock                        = &sync.Mutex{}
	instance *ConfigServiceImpl = nil
)

type ConfigServiceImpl struct {
	viper *viper.Viper
}

// Get one and only ConfigServiceImpl instance.
func GetConfigService() *ConfigServiceImpl {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = newConfigService()
		}
	}
	return instance
}

// Creates a new ConfigServiceImpl object with an initialized Viper object, and
// reads the configuration file from the path specified in the GARAGESERVICE_CONFIG_PATH.
// If the path is not set, it will default to the current directory.
func newConfigService() *ConfigServiceImpl {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	configPath := os.Getenv("GARAGESERVICE_CONFIG_PATH")
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("config: fatal error while parsing config file: %s", err))
	}
	return &ConfigServiceImpl{viper: v}
}

// Verifies that all mandatory keys are set in the configuration file,
// and that no unknown keys are present.
func (c ConfigServiceImpl) verifyKeys() error {
	for key, mandatory := range KNOWN_KEYS {
		if mandatory && !c.viper.IsSet(key) {
			return fmt.Errorf("config: configuration property %s is mandatory", key)
		}
	}
	for _, key := range c.viper.AllKeys() {
		if _, found := KNOWN_KEYS[key]; !found {
			return fmt.Errorf("config: configuration property %s is unknown", key)
		}
	}
	return nil
}

// Verify that the configuration is valid.
func (c ConfigServiceImpl) Verify() error {
	if err := c.verifyKeys(); err != nil {
		return err
	}
	mode := c.viper.GetString("mode")
	if mode != "development" && mode != "production" {
		return fmt.Errorf("config: mode must be either 'development' or 'production'")
	}
	port := c.viper.GetInt("bind.port")
	if port < 0 || port > 65535 {
		return fmt.Errorf("config: bind.port must be a valid port number")
	}
	togglePin := c.viper.GetInt("gpio.toggle_pin")
	if togglePin < 0 {
		return fmt.Errorf("config: gpio.toggle_pin must be a valid pin number")
	}
	openPin := c.viper.GetInt("gpio.open_pin")
	if openPin < 0 {
		return fmt.Errorf("config: gpio.open_pin must be a valid pin number")
	}
	closedPin := c.viper.GetInt("gpio.closed_pin")
	if closedPin < 0 {
		return fmt.Errorf("config: gpio.closed_pin must be a valid pin number")
	}
	apiKeys := c.viper.GetStringSlice("api_keys")
	if len(apiKeys) == 0 {
		return fmt.Errorf("config: api_keys must contain at least one key")
	}

	return nil
}

func (c ConfigServiceImpl) GetMode() string {
	return c.viper.GetString("mode")
}

func (c ConfigServiceImpl) GetBindPort() int {
	return c.viper.GetInt("bind.port")
}

func (c ConfigServiceImpl) GetBindHost() string {
	return c.viper.GetString("bind.host")
}

func (c ConfigServiceImpl) GetTogglePin() int {
	return c.viper.GetInt("gpio.toggle_pin")
}

func (c ConfigServiceImpl) GetOpenPin() int {
	return c.viper.GetInt("gpio.open_pin")
}

func (c ConfigServiceImpl) GetClosedPin() int {
	return c.viper.GetInt("gpio.closed_pin")
}

func (c ConfigServiceImpl) GetApiKeys() []string {
	return c.viper.GetStringSlice("api_keys")
}
