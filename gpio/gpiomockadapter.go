package gpio

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type GPIOMockAdapter struct {
	togglePin      int
	openPin        int
	closedPin      int
	togglePinState bool
	openState      bool
	closedState    bool
}

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

func (g *GPIOMockAdapter) WriteTogglePin(value bool) error {
	log.Info().Msg(fmt.Sprintf("Mock GPIO: Writing to pin %d: %v", g.togglePin, value))
	if !g.togglePinState && value {
		log.Info().Msg("Mock GPIO: Toggling garage door")
		g.openState = !g.openState
		g.closedState = !g.closedState
	}
	g.togglePinState = value
	return nil
}

func (g *GPIOMockAdapter) ReadOpenPin() (bool, error) {
	log.Info().Msg(fmt.Sprintf("Mock GPIO: Reading from pin %d: %v", g.openPin, g.openState))
	return g.openState, nil
}

func (g *GPIOMockAdapter) ReadClosedPin() (bool, error) {
	log.Info().Msg(fmt.Sprintf("Mock GPIO: Reading from pin %d: %v", g.closedPin, g.closedState))
	return g.closedState, nil
}
