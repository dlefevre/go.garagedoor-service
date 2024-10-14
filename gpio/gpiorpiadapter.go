package gpio

import (
	"fmt"
	"sync"

	"github.com/stianeikeland/go-rpio/v4"
)

var once sync.Once

// GPIORPiAdapter is an adapter for the Raspberry Pi GPIO pins.
type GPIORPiAdapter struct {
	togglePin rpio.Pin
	openPin   rpio.Pin
	closedPin rpio.Pin
}

// NewGPIORPiAdapter creates a new GPIORPiAdapter.
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

// WriteTogglePin sets the toggle pin to high when value is true.
func (g *GPIORPiAdapter) WriteTogglePin(value bool) {
	if value {
		g.togglePin.High()
	} else {
		g.togglePin.Low()
	}
}

// ReadOpenPin returns true if the open pin is high, and false otherwise
func (g *GPIORPiAdapter) ReadOpenPin() bool {
	state := rpio.ReadPin(g.openPin)
	return state == rpio.High
}

// ReadClosedPin returns true if the closed pin is high, and false otherwise
func (g *GPIORPiAdapter) ReadClosedPin() bool {
	state := rpio.ReadPin(g.closedPin)
	return state == rpio.High
}

// Reset isn't implemented dfor this adapter type.
func (g *GPIORPiAdapter) Reset() error {
	return fmt.Errorf("the `Reset` function isn't implemented for the GPIORPiAdapter")
}
