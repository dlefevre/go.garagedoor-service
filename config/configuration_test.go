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
	if err := Verify(); err != nil {
		t.Fatalf("Error verifying configuration: %v", err)
	}
}

func TestMode(t *testing.T) {
	if GetMode() != "development" {
		t.Fatalf("Expected mode to be development, got %s", GetMode())
	}
}

func TestBindPortAndHost(t *testing.T) {
	if GetBindHost() != "127.0.0.1" {
		t.Fatalf("Expected bind  host to be 127.0.0.1, got %s", GetBindHost())
	}
	if GetBindPort() != 8000 {
		t.Fatalf("Expected bind port to be 8080, got %d", GetBindPort())
	}
}

func TestGPIO(t *testing.T) {
	if GetTogglePin() != 11 {
		t.Fatalf("Expected toggle pin to be 11, got %d", GetTogglePin())
	}
	if GetOpenPin() != 21 {
		t.Fatalf("Expected open pin to be 21, got %d", GetOpenPin())
	}
	if GetClosedPin() != 22 {
		t.Fatalf("Expected closed pin to be 22, got %d", GetClosedPin())
	}
}

func TestAPIKeys(t *testing.T) {
	keys := GetAPIKeys()
	if len(keys) != 1 {
		t.Fatalf("Expected 1 API key, got %d", len(keys))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(keys[0]), []byte("test")); err != nil {
		t.Fatalf("Expected API key to be bcrypt digest of 'test'")
	}
}

func TestMQTT(t *testing.T) {
	if !GetMQTTEnabled() {
		t.Fatalf("Expected MQTT to be enabled")
	}
	if GetMQTTURL() != "mqtt://mqtt.eclipseprojects.io:1883" {
		t.Fatalf("Expected MQTT URL to be mqtt://mqtt.eclipseprojects.io:1883, got %s", GetMQTTURL())
	}
	if GetMQTTClientID() != "garage_door" {
		t.Fatalf("Expected MQTT client ID to be garage_door, got %s", GetMQTTClientID())
	}
	if GetMQTTTopicPrefix() != "/homeassistant/" {
		t.Fatalf("Expected MQTT topic prefix to be /homeassistant/, got %s", GetMQTTTopicPrefix())
	}
}
