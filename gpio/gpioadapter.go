package gpio

import "github.com/dlefevre/go.garagedoor-service/config"

type GPIOAdapter interface {
	WriteTogglePin(value bool) error
	ReadOpenPin() (bool, error)
	ReadClosedPin() (bool, error)
	Reset() error
}

func GetGPIOAdapter() GPIOAdapter {
	switch config.GetMode() {
	case "prod":
		panic("Production mode not supported yet")
	case "development":
		return NewGPIOMockAdapter(config.GetTogglePin(), config.GetOpenPin(), config.GetClosedPin())
	default:
		panic("Unknown mode")
	}
}
