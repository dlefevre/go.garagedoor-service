package gpio

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// GPIOMockAdapter is a mock GPIO adapter, which:
// - mimicks the behavior of the garage door, without the delays of a physical door and motor.
// - reports all actions to the log.
type GPIOMockAdapter struct {
	togglePin      int
	openPin        int
	closedPin      int
	togglePinState bool
	openState      bool
	closedState    bool
}

// NewGPIOMockAdapter creates a new GPIOMockAdapter.
func NewGPIOMockAdapter(togglePin int, openPin int, closedPin int) *GPIOMockAdapter {
	log.Info().Msg("Mock GPIO: Creating mock GPIO adapter")
	return &GPIOMockAdapter{
		togglePin:   togglePin,
		openPin:     openPin,
		closedPin:   closedPin,
		openState:   false,
		closedState: true,
	}
}

// WriteTogglePin sets the toggle pin to high when value is true.
func (g *GPIOMockAdapter) WriteTogglePin(value bool) {
	log.Info().Msg(fmt.Sprintf("Mock GPIO: Writing to pin %d: %v", g.togglePin, value))
	if !g.togglePinState && value {
		log.Info().Msg("Mock GPIO: Toggling garage door")
		g.openState = !g.openState
		g.closedState = !g.closedState
	}
	g.togglePinState = value
}

// ReadOpenPin returns true if the open pin is high, and false otherwise.
func (g *GPIOMockAdapter) ReadOpenPin() bool {
	log.Info().Msg(fmt.Sprintf("Mock GPIO: Reading from pin %d: %v", g.openPin, g.openState))
	return g.openState
}

// ReadClosedPin returns true if the closed pin is high, and false otherwise.
func (g *GPIOMockAdapter) ReadClosedPin() bool {
	log.Info().Msg(fmt.Sprintf("Mock GPIO: Reading from pin %d: %v", g.closedPin, g.closedState))
	return g.closedState
}

// Reset the pins to their initial state.
func (g *GPIOMockAdapter) Reset() error {
	log.Info().Msg("Mock GPIO: Resetting pins")
	g.openState = false
	g.closedState = true
	return nil
}
