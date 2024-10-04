package config

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func init() {
	// Set the environment variable for the configuration path
	os.Setenv("GARAGESERVICE_CONFIG_PATH", "..")
}

func TestVerify(t *testing.T) {
	config := GetConfigService()
	if err := config.Verify(); err != nil {
		t.Fatalf("Error verifying configuration: %v", err)
	}
}

func TestMode(t *testing.T) {
	config := GetConfigService()
	if config.GetMode() != "development" {
		t.Fatalf("Expected mode to be development, got %s", config.GetMode())
	}
}

func TestBindPortAndHost(t *testing.T) {
	config := GetConfigService()
	if config.GetBindHost() != "127.0.0.1" {
		t.Fatalf("Expected bind  host to be 127.0.0.1, got %s", config.GetBindHost())
	}
	if config.GetBindPort() != 8000 {
		t.Fatalf("Expected bind port to be 8080, got %d", config.GetBindPort())
	}
}

func TestGPIO(t *testing.T) {
	config := GetConfigService()
	if config.GetTogglePin() != 11 {
		t.Fatalf("Expected toggle pin to be 11, got %d", config.GetTogglePin())
	}
	if config.GetOpenPin() != 21 {
		t.Fatalf("Expected open pin to be 21, got %d", config.GetOpenPin())
	}
	if config.GetClosedPin() != 22 {
		t.Fatalf("Expected closed pin to be 22, got %d", config.GetClosedPin())
	}
}

func TestAPIKeys(t *testing.T) {
	config := GetConfigService()
	keys := config.GetApiKeys()
	if len(keys) != 1 {
		t.Fatalf("Expected 1 API key, got %d", len(keys))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(keys[0]), []byte("test")); err != nil {
		t.Fatalf("Expected API key to be bcrypt digest of 'test'")
	}
}
