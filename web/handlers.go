package web

import (
	"encoding/json"
	"net/http"

	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

// SimpleResponse is a simple response object, containing a result (ok, nok).
type SimpleResponse struct {
	Result string `json:"result"`
}

// ErrorResponse is a response object for errors, containing  a result (nok) and a message.
type ErrorResponse struct {
	SimpleResponse
	Message string `json:"message"`
}

// StateResponse is a response object for the state of the door, containing a result (ok) and the state.
type StateResponse struct {
	SimpleResponse
	State string `json:"state"`
}

// CommandMessage is a message object for commands, containing a command.
type CommandMessage struct {
	Command string `json:"command"`
}

// Health check handler.
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, SimpleResponse{
		Result: "OK",
	})
}

// Toggle forwards a toggle request to the DoorControllerService.
func toggle(c echo.Context) error {
	dc := controller.GetDoorControllerService()
	dc.RequestToggle()
	return c.JSON(http.StatusOK, SimpleResponse{
		Result: "ok",
	})
}

// Get the current state of the door.
func state(c echo.Context) error {
	dc := controller.GetDoorControllerService()
	return c.JSON(http.StatusOK, StateResponse{
		SimpleResponse: SimpleResponse{
			Result: "ok",
		},
		State: dc.GetStateStr(),
	})
}

// Handler for the websocket.
func ws(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// Add a state listener to send state updates to the websocket.
		WebSocketStateListener := &WebSocketStateListener{}
		WebSocketStateListener.Connect(ws)
		defer WebSocketStateListener.Disconnect()

		// Read messages from the websocket.
		dc := controller.GetDoorControllerService()
		for {
			var msg []byte
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				log.Error().Msgf("Error reading message from websocket: %v", err)
				break
			}
			var command CommandMessage
			if err := json.Unmarshal(msg, &command); err != nil {
				log.Error().Msgf("Error parsing message from websocket: %v", err)
				break
			}
			switch command.Command {
			case "toggle":
				dc.RequestToggle()
			case "state":
				dc.RequestState()
			default:
				log.Warn().Msgf("Unknown command: %s", command.Command)
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
