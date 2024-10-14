package gpio

import (
	"fmt"
	"sync"

	"github.com/stianeikeland/go-rpio/v4"
)

var once sync.Once

type GPIORPiAdapter struct {
	togglePin rpio.Pin
	openPin   rpio.Pin
	closedPin rpio.Pin
}

// Create a new GPIORPiAdapter.
func NewGPIORPiAdapter(togglePin int, openPin int, closedPin int) *GPIORPiAdapter {
	once.Do(func() {
		rpio.Open()
	})

	adapter := &GPIORPiAdapter{
		togglePin: rpio.Pin(togglePin),
		openPin:   rpio.Pin(openPin),
		closedPin: rpio.Pin(closedPin),
	}
	adapter.togglePin.Output()
	adapter.openPin.Input()
	adapter.closedPin.Input()

	return adapter
}

// Set the toggle pin to high or low.
func (g *GPIORPiAdapter) WriteTogglePin(value bool) {
	if value {
		g.togglePin.High()
	} else {
		g.togglePin.Low()
	}
}

// Read the state of the open pin.
func (g *GPIORPiAdapter) ReadOpenPin() bool {
	state := rpio.ReadPin(g.openPin)
	return state == rpio.High
}

// Read the state of the closed pin.
func (g *GPIORPiAdapter) ReadClosedPin() bool {
	state := rpio.ReadPin(g.closedPin)
	return state == rpio.High
}

// Reset isn't implemented dfor this adapter type.
func (g *GPIORPiAdapter) Reset() error {
	return fmt.Errorf("the `Reset` function isn't implemented for the GPIORPiAdapter")
}
