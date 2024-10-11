package web

import (
	"os"
	"testing"
	"time"
)

func init() {
	// Set the environment variable for the configuration path
	os.Setenv("GARAGESERVICE_CONFIG_PATH", "..")
}

func TestStartStop(t *testing.T) {
	ws := GetWebService()
	ws.Start()
	time.Sleep(1 * time.Second)
	ws.Stop()
	time.Sleep(1 * time.Second)
	ws.Start()
	time.Sleep(1 * time.Second)
	ws.Stop()
}
