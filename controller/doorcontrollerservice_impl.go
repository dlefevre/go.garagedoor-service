package controller

import (
	"sync"
	"time"

	"github.com/dlefevre/go.garagedoor-service/gpio"
	"github.com/rs/zerolog/log"
)

// Enum pseudo-type.
type Enum uint

// Queue size for the command channel.
const queueSize = 10

// Enumeration of commands.
const (
	CmdDummy  Enum = iota // CmdDummy does nothing, but prevents errors when closing the channel.
	CmdToggle             // CmdToggle identifies the toggle command
	CmdState              // CmdState identifies the state request command
)

// Enumeration of states.
const (
	StateOpen    Enum = iota // StateOpen represents the open state
	StateClosed              // StateClosed represents the closed state
	StateUnknown             // StateUnknown represents the unknown state
)

var (
	instance *DoorControllerServiceImpl
	once     sync.Once
)

// DoorControllerServiceImpl implements the service for controlling the garagedoor and reporting its state.
type DoorControllerServiceImpl struct {
	command        chan Enum
	stateListeners []func(string)
	state          Enum
	lock           sync.RWMutex
	adapter        gpio.GPIOAdapter
	wg             sync.WaitGroup
	running        bool
}

// GetDoorControllerService returns the one and only DoorControllerServiceImpl instance.
func GetDoorControllerService() *DoorControllerServiceImpl {
	once.Do(func() {
		instance = newDoorControllerService()
	})
	return instance
}

// Creates a new DoorControllerServiceImpl object.
func newDoorControllerService() *DoorControllerServiceImpl {
	return &DoorControllerServiceImpl{
		command:        nil,
		stateListeners: make([]func(string), 0),
		state:          StateUnknown,
		lock:           sync.RWMutex{},
		adapter:        gpio.GetGPIOAdapter(),
		wg:             sync.WaitGroup{},
		running:        false,
	}
}

// Reset GPIO Adapter and state.
func (d *DoorControllerServiceImpl) Reset() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.adapter.Reset()
	d.state = StateUnknown
}

// Main loop for handling commands.
func (d *DoorControllerServiceImpl) commandLoop() {
	defer d.wg.Done()

	for d.running {
		switch <-d.command {
		case CmdToggle:
			d.toggle()
		case CmdState:
			d.broadcastState()
		case CmdDummy:
			// Do nothing
		default:
			log.Warn().Msgf("unknown command: %v", d.command)
		}
	}

	log.Info().Msg("commandLoop exiting")
}

// Main loop for reading and broadcasting the state of the garagedoor.
func (d *DoorControllerServiceImpl) stateLoop() {
	defer d.wg.Done()

	for d.running {
		state := d.readCurrentState()
		if d.stateDiffers(state) {
			d.setState(state)
			d.broadcastState()
		}

		time.Sleep(250 * time.Millisecond)
	}

	log.Info().Msg("stateLoop exiting")
}

// Start all goroutines.
func (d *DoorControllerServiceImpl) Start() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.running = true
	d.command = make(chan Enum, queueSize)
	go d.commandLoop()
	go d.stateLoop()
	d.wg.Add(2)
}

// Stop all goroutines, gracefully.
func (d *DoorControllerServiceImpl) Stop() {
	d.lock.Lock()
	d.running = false
	close(d.command)
	d.lock.Unlock()
	log.Info().Msg("Stopping DoorControllerService")

	d.wg.Wait()
	log.Info().Msg("DoorControllerService stopped")
}

// RequestToggle puts a toggle command on the command queue
func (d *DoorControllerServiceImpl) RequestToggle() {
	d.command <- CmdToggle
}

// RequestState puts a state update request on the command queue
func (d *DoorControllerServiceImpl) RequestState() {
	d.command <- CmdState
}

// GetStateStr returns a string representation of the current state.
func (d *DoorControllerServiceImpl) GetStateStr() string {
	return d.stateStr()
}

// AddStateListener adds a listerer for state changes. When added, no initial state is sent.
// If an update is needed, RequestState() should be called.
// Returns an index that can be used to remove the listener.
func (d *DoorControllerServiceImpl) AddStateListener(handler func(string)) uint {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.stateListeners = append(d.stateListeners, handler)
	return uint(len(d.stateListeners) - 1)
}

// RemoveStateListener removes a listener by index.
func (d *DoorControllerServiceImpl) RemoveStateListener(index uint) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.stateListeners = append(d.stateListeners[:index], d.stateListeners[index+1:]...)
}

// Generate a string representation of the current state.
func (d *DoorControllerServiceImpl) stateStr() string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	switch d.state {
	case StateOpen:
		return "open"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}

// Check if the object's state differs from the given state.
func (d *DoorControllerServiceImpl) stateDiffers(state Enum) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.state != state
}

// Set the object's state to the given state.
func (d *DoorControllerServiceImpl) setState(state Enum) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.state = state
}

// Broadcast the current state to all listeners.
func (d *DoorControllerServiceImpl) broadcastState() {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, listener := range d.stateListeners {
		listener(d.stateStr())
	}
}

// Toggle the garagedoor.
func (d *DoorControllerServiceImpl) toggle() {
	d.adapter.WriteTogglePin(true)
	time.Sleep(250 * time.Millisecond)
	d.adapter.WriteTogglePin(false)
	time.Sleep(250 * time.Millisecond)
}

// Read the current state from the two pins connected to the magnetic switches.
func (d *DoorControllerServiceImpl) readCurrentState() Enum {
	open := d.adapter.ReadOpenPin()
	closed := d.adapter.ReadClosedPin()

	if open && !closed {
		return StateOpen
	} else if !open && closed {
		return StateClosed
	} else {
		return StateUnknown
	}
}
