package gpio

type GPIOAdapter interface {
	writeTogglePin(value bool) error
	readOpenPin() (bool, error)
	readClosedPin() (bool, error)
}
