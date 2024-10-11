package gpio

import (
	"os"
	"testing"

	"github.com/dlefevre/go.garagedoor-service/config"

	"github.com/rs/zerolog"
)

func init() {
	os.Setenv("GARAGESERVICE_CONFIG_PATH", "..")
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func TestInitialState(t *testing.T) {
	gpio := NewGPIOMockAdapter(config.GetTogglePin(), config.GetOpenPin(), config.GetClosedPin())
	if open, err := gpio.ReadOpenPin(); err != nil || open {
		t.Fatalf("Expected open pin to be false, got %v", open)
	}
	if closed, err := gpio.ReadClosedPin(); err != nil || !closed {
		t.Fatalf("Expected closed pin to be true, got %v", closed)
	}
}

func toggleHelper(t *testing.T, gpio *GPIOMockAdapter, expectedOpen bool, expectedClosed bool) {
	if err := gpio.WriteTogglePin(true); err != nil {
		t.Fatalf("Error writing to toggle pin: %v", err)
	}
	if err := gpio.WriteTogglePin(false); err != nil {
		t.Fatalf("Error writing to toggle pin: %v", err)
	}
	if open, err := gpio.ReadOpenPin(); err != nil || open != expectedOpen {
		t.Fatalf("Expected open pin to be %v, got %v", expectedOpen, open)
	}
	if closed, err := gpio.ReadClosedPin(); err != nil || closed != expectedClosed {
		t.Fatalf("Expected closed pin to be %v, got %v", expectedClosed, closed)
	}
}

func TestToggle(t *testing.T) {
	gpio := NewGPIOMockAdapter(config.GetTogglePin(), config.GetOpenPin(), config.GetClosedPin())
	toggleHelper(t, gpio, true, false)
	toggleHelper(t, gpio, false, true)
}

func TestReset(t *testing.T) {
	gpio := NewGPIOMockAdapter(config.GetTogglePin(), config.GetOpenPin(), config.GetClosedPin())
	toggleHelper(t, gpio, true, false)
	if err := gpio.Reset(); err != nil {
		t.Fatalf("Error resetting pins: %v", err)
	}
	if open, err := gpio.ReadOpenPin(); err != nil || open {
		t.Fatalf("Expected open pin to be false, got %v", open)
	}
	if closed, err := gpio.ReadClosedPin(); err != nil || !closed {
		t.Fatalf("Expected closed pin to be true, got %v", closed)
	}
}
