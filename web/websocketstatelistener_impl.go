package web

import (
	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

// WebSocketStateListener implements the handler to report state changes to a websocket, and register and unregister
// the listener with the DoorControllerService.
type WebSocketStateListener struct {
	ws    *websocket.Conn
	index uint
}

// StateChange handles sending state updates to the websocket.
func (w *WebSocketStateListener) StateChanged(state string) {
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

// Connect registers the websocket, and adds a state listener to send state updates to the websocket.
func (w *WebSocketStateListener) Connect(ws *websocket.Conn) {
	w.ws = ws
	dc := controller.GetDoorControllerService()
	w.index = dc.AddStateListener(w.StateChanged)
}

// Disconnect removes the state listener.
func (w *WebSocketStateListener) Disconnect() {
	dc := controller.GetDoorControllerService()
	dc.RemoveStateListener(w.index)
}
