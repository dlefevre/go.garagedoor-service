package controller

import (
	"sync"
	"time"

	"github.com/dlefevre/go.garagedoor-service/gpio"
	"github.com/rs/zerolog/log"
)

type Enum uint

// Queue size for the command channel.
const QUEUE_SIZE = 10

// Enumeration of commands.
const (
	CMD_DUMMY Enum = iota
	CMD_TOGGLE
	CMD_STATE
)

// Enumeration of states.
const (
	STATE_OPEN Enum = iota
	STATE_CLOSED
	STATE_UNKNOWN
)

var (
	instance *DoorControllerServiceImpl = nil
	lock                                = &sync.Mutex{}
)

type DoorControllerServiceImpl struct {
	command        chan Enum
	stateListeners []func(string)
	state          Enum
	lock           sync.RWMutex
	adapter        gpio.GPIOAdapter
	wg             sync.WaitGroup
	running        bool
}

// Get one and only DoorControllerServiceImpl instance.
func GetDoorControllerService() *DoorControllerServiceImpl {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = newDoorControllerService()
		}
	}
	return instance
}

// Creates a new DoorControllerServiceImpl object.
func newDoorControllerService() *DoorControllerServiceImpl {
	return &DoorControllerServiceImpl{
		command:        nil,
		stateListeners: make([]func(string), 0),
		state:          STATE_UNKNOWN,
		lock:           sync.RWMutex{},
		adapter:        gpio.GetGPIOAdapter(),
		wg:             sync.WaitGroup{},
		running:        false,
	}
}

func (d *DoorControllerServiceImpl) commandLoop() {
	defer d.wg.Done()

	for d.running {
		switch <-d.command {
		case CMD_TOGGLE:
			d.toggle()
		case CMD_STATE:
			d.broadcastState()
		case CMD_DUMMY:
			// Do nothing
		default:
			log.Warn().Msgf("unknown command: %v", d.command)
		}
	}

	log.Info().Msg("commandLoop exiting")
}

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

func (d *DoorControllerServiceImpl) Start() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.running = true
	d.command = make(chan Enum, QUEUE_SIZE)
	go d.commandLoop()
	go d.stateLoop()
	d.wg.Add(2)
}

func (d *DoorControllerServiceImpl) Stop() {
	d.lock.Lock()
	d.running = false
	close(d.command)
	d.lock.Unlock()
	log.Info().Msg("Stopping DoorControllerService")

	d.wg.Wait()
	log.Info().Msg("DoorControllerService stopped")
}

func (d *DoorControllerServiceImpl) RequestToggle() {
	d.command <- CMD_TOGGLE
}

func (d *DoorControllerServiceImpl) RequestState() {
	d.command <- CMD_STATE
}

func (d *DoorControllerServiceImpl) GetStateStr() string {
	return d.stateStr()
}

func (d *DoorControllerServiceImpl) AddStateListener(handler func(string)) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.stateListeners = append(d.stateListeners, handler)
}

func (d *DoorControllerServiceImpl) stateStr() string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	switch d.state {
	case STATE_OPEN:
		return "open"
	case STATE_CLOSED:
		return "closed"
	default:
		return "unknown"
	}
}

func (d *DoorControllerServiceImpl) stateDiffers(state Enum) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.state != state
}

func (d *DoorControllerServiceImpl) setState(state Enum) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.state = state
}

func (d *DoorControllerServiceImpl) broadcastState() {
	d.lock.RLock()
	defer d.lock.RUnlock()
	for _, listener := range d.stateListeners {
		listener(d.stateStr())
	}
}

func (d *DoorControllerServiceImpl) toggle() {
	d.adapter.WriteTogglePin(true)
	time.Sleep(250 * time.Millisecond)
	d.adapter.WriteTogglePin(false)
	time.Sleep(250 * time.Millisecond)
}

func (d *DoorControllerServiceImpl) readCurrentState() Enum {
	open, err := d.adapter.ReadOpenPin()
	if err != nil {
		log.Fatal().Msgf("Error reading open pin: %v", err)
		panic(err)
	}
	closed, err := d.adapter.ReadClosedPin()
	if err != nil {
		log.Fatal().Msgf("Error reading closed pin: %v", err)
		panic(err)
	}

	if open && !closed {
		return STATE_OPEN
	} else if !open && closed {
		return STATE_CLOSED
	} else {
		return STATE_UNKNOWN
	}
}
