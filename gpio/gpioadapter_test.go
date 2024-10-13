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
	if open := gpio.ReadOpenPin(); open {
		t.Fatalf("Expected open pin to be false, got %v", open)
	}
	if closed := gpio.ReadClosedPin(); !closed {
		t.Fatalf("Expected closed pin to be true, got %v", closed)
	}
}

func toggleHelper(t *testing.T, gpio *GPIOMockAdapter, expectedOpen bool, expectedClosed bool) {
	gpio.WriteTogglePin(true)
	gpio.WriteTogglePin(false)
	if open := gpio.ReadOpenPin(); open != expectedOpen {
		t.Fatalf("Expected open pin to be %v, got %v", expectedOpen, open)
	}
	if closed := gpio.ReadClosedPin(); closed != expectedClosed {
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
	gpio.Reset()
	if open := gpio.ReadOpenPin(); open {
		t.Fatalf("Expected open pin to be false, got %v", open)
	}
	if closed := gpio.ReadClosedPin(); !closed {
		t.Fatalf("Expected closed pin to be true, got %v", closed)
	}
}
