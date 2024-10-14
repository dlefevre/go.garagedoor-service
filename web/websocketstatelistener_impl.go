package web

import (
	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

type WebSocketStateListerer struct {
	ws    *websocket.Conn
	index uint
}

// Handler for sending state updates to the websocket.
func (w *WebSocketStateListerer) StateChanged(state string) {
	err := websocket.JSON.Send(w.ws, StateResponse{
		SimpleResponse: SimpleResponse{
			Result: "ok",
		},
		State: state,
	})
	if err != nil {
		log.Error().Msgf("Error sending state to websocket: %v", err)
	}
}

// Register the websocket, and add a state listener to send state updates to the websocket.
func (w *WebSocketStateListerer) Connect(ws *websocket.Conn) {
	w.ws = ws
	dc := controller.GetDoorControllerService()
	w.index = dc.AddStateListener(w.StateChanged)
}

// Remove the state listener.
func (w *WebSocketStateListerer) Disconnect() {
	dc := controller.GetDoorControllerService()
	dc.RemoveStateListener(w.index)
}
