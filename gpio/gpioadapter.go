package gpio

import "github.com/dlefevre/go.garagedoor-service/config"

type GPIOAdapter interface {
	WriteTogglePin(value bool)
	ReadOpenPin() bool
	ReadClosedPin() bool
	Reset() error
}

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
