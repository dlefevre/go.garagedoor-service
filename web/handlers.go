package web

import (
	"encoding/json"
	"net/http"

	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

type SimpleResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	SimpleResponse
	Message string `json:"message"`
}

type StateResponse struct {
	SimpleResponse
	State string `json:"state"`
}

type CommandMessage struct {
	Command string `json:"command"`
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, SimpleResponse{
		Result: "OK",
	})
}

func toggle(c echo.Context) error {
	dc := controller.GetDoorControllerService()
	dc.RequestToggle()
	return c.JSON(http.StatusOK, SimpleResponse{
		Result: "ok",
	})
}

func state(c echo.Context) error {
	dc := controller.GetDoorControllerService()
	return c.JSON(http.StatusOK, StateResponse{
		SimpleResponse: SimpleResponse{
			Result: "ok",
		},
		State: dc.GetStateStr(),
	})
}

func ws(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// Add a state listener to send state updates to the websocket.
		WebSocketStateListerer := &WebSocketStateListerer{}
		WebSocketStateListerer.Connect(ws)
		defer WebSocketStateListerer.Disconnect()

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
