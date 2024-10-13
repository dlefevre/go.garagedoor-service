package web

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

func init() {
	// Set the environment variable for the configuration path
	os.Setenv("GARAGESERVICE_CONFIG_PATH", "..")
}

func setup() {
	dc := controller.GetDoorControllerService()
	dc.Start()
	ws := GetWebService()
	ws.Start()
	dc.Reset()

	time.Sleep(500 * time.Microsecond)
}

func teardown() {
	dc := controller.GetDoorControllerService()
	dc.Stop()
	ws := GetWebService()
	ws.Stop()
}

func toggleHelper(t *testing.T) {
	client := &http.Client{}

	req, err := http.NewRequest("POST", "http://localhost:8000/toggle", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	req.Header.Add("x-api-key", "test")
	log.Debug().Msg("Sending request")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response: %v", err)
	}
	var myResponse SimpleResponse
	if err := json.Unmarshal(body, &myResponse); err != nil {
		t.Fatalf("Error unmarshalling response: %v", err)
	}
	if myResponse.Result != "ok" {
		t.Fatalf("Expected result to be ok, got %s", myResponse.Result)
	}
}

func stateHelper(t *testing.T, expectedState string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://localhost:8000/state", nil)
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	req.Header.Add("x-api-key", "test")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Error reading response: %v", err)
	}
	var myResponse StateResponse
	if err := json.Unmarshal(body, &myResponse); err != nil {
		t.Fatalf("Error unmarshalling response: %v", err)
	}
	if myResponse.Result != "ok" {
		t.Fatalf("Expected result to be ok, got %s", myResponse.Result)
	}
	if myResponse.State != expectedState {
		t.Fatalf("Expected state to be %s, got %s", expectedState, myResponse.State)
	}
}

func TestStartStop(t *testing.T) {
	setup()
	teardown()
}

func TestToggle(t *testing.T) {
	setup()
	defer teardown()
	toggleHelper(t)
}

func TestState(t *testing.T) {
	setup()
	defer teardown()
	stateHelper(t, "closed")
	toggleHelper(t)
	time.Sleep(500 * time.Millisecond)
	stateHelper(t, "open")
	toggleHelper(t)
	time.Sleep(500 * time.Millisecond)
	stateHelper(t, "closed")
}

func TestWebSocket(t *testing.T) {
	setup()
	defer teardown()

	// Connect
	config, err := websocket.NewConfig("ws://localhost:8000/ws", "http://localhost:8000")
	if err != nil {
		t.Fatalf("Error creating websocket config: %v", err)
	}
	config.Header = http.Header{
		"x-api-key": []string{"test"},
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {
		t.Fatalf("Error connecting to websocket: %v", err)
	}
	defer ws.Close()

	// Listen for responses
	stateStr := ""
	go func() {
		for {
			var msg []byte
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				break
			}
			var myResponse StateResponse
			if err := json.Unmarshal(msg, &myResponse); err != nil {
				log.Fatal().Msgf("Error unmarshalling response: %v", err)
			}
			if myResponse.Result != "ok" {
				log.Fatal().Msgf("Expected result to be ok, got %s", myResponse.Result)
			}
			stateStr = myResponse.State
		}
	}()

	cmdState := CommandMessage{
		Command: "state",
	}
	cmdToggle := CommandMessage{
		Command: "toggle",
	}
	if err := websocket.JSON.Send(ws, cmdState); err != nil {
		t.Fatalf("Error sending state command: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	if stateStr != "closed" {
		t.Fatalf("Expected state to be closed, got %s", stateStr)
	}
	if err := websocket.JSON.Send(ws, cmdToggle); err != nil {
		t.Fatalf("Error sending toggle command: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	if stateStr != "open" {
		t.Fatalf("Expected state to be open, got %s", stateStr)
	}
	if err := websocket.JSON.Send(ws, cmdToggle); err != nil {
		t.Fatalf("Error sending toggle command: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	if stateStr != "closed" {
		t.Fatalf("Expected state to be closed, got %s", stateStr)
	}
}
