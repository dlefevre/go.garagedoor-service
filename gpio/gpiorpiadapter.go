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

func (g *GPIORPiAdapter) WriteTogglePin(value bool) {
	if value {
		g.togglePin.High()
	} else {
		g.togglePin.Low()
	}
}

func (g *GPIORPiAdapter) ReadOpenPin() bool {
	state := rpio.ReadPin(g.openPin)
	return state == rpio.High
}

func (g *GPIORPiAdapter) ReadClosedPin() bool {
	state := rpio.ReadPin(g.closedPin)
	return state == rpio.High
}

func (g *GPIORPiAdapter) Reset() error {
	return fmt.Errorf("the `Reset` function isn't implemented for the GPIORPiAdapter")
}
