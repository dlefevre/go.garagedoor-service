package gpio

import "github.com/dlefevre/go.garagedoor-service/config"

// GPIOAdapter specifies the interface for GPIO operations.
type GPIOAdapter interface {
	WriteTogglePin(value bool)
	ReadOpenPin() bool
	ReadClosedPin() bool
	Reset() error
}

// GetGPIOAdapter returns the GPIO adapter based on the current mode.
func GetGPIOAdapter() GPIOAdapter {
	switch config.GetMode() {
	case "production":
		return NewGPIORPiAdapter(config.GetTogglePin(), config.GetOpenPin(), config.GetClosedPin())
	case "development":
		return NewGPIOMockAdapter(config.GetTogglePin(), config.GetOpenPin(), config.GetClosedPin())
	default:
		panic("Unknown mode")
	}
}
