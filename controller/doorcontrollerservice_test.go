package controller

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func init() {
	// Set the environment variable for the configuration path
	os.Setenv("GARAGESERVICE_CONFIG_PATH", "..")
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

func TestCreateDoorController(t *testing.T) {
	controller := GetDoorControllerService()
	if controller == nil {
		t.Fatalf("Expected controller to be created")
	}
}

func TestStartStop(t *testing.T) {
	controller := GetDoorControllerService()
	controller.Reset()
	controller.Start()
	time.Sleep(1 * time.Second)
	controller.Stop()
}

func TestRestart(t *testing.T) {
	controller := GetDoorControllerService()
	controller.Reset()
	controller.Start()
	time.Sleep(1 * time.Second)
	controller.Stop()
	controller.Start()
	time.Sleep(1 * time.Second)
	controller.Stop()
}

func TestToggle(t *testing.T) {
	controller := GetDoorControllerService()
	controller.Reset()
	controller.Start()

	time.Sleep(1 * time.Second)
	if controller.GetStateStr() != "closed" {
		t.Fatalf("Expected state to be closed, got %s", controller.GetStateStr())
	}

	controller.RequestToggle()
	time.Sleep(1 * time.Second)
	if controller.GetStateStr() != "open" {
		t.Fatalf("Expected state to be open, got %s", controller.GetStateStr())
	}

	controller.Stop()
}

func TestListener(t *testing.T) {
	controller := GetDoorControllerService()
	controller.Reset()
	controller.Start()

	state := ""
	controller.AddStateListener(func(s string) {
		state = s
	})

	controller.RequestState()
	time.Sleep(1 * time.Second)
	if state != "closed" {
		t.Fatalf("Expected state to be closed, got %s", state)
	}

	controller.RequestToggle()
	time.Sleep(1 * time.Second)
	if state != "open" {
		t.Fatalf("Expected state to be open, got %s", state)
	}

	controller.Stop()
}
