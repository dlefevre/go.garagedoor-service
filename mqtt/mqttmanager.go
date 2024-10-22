package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/dlefevre/go.garagedoor-service/config"
	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	instance *MQTTManager
	once     sync.Once
)

// MQTTManager is a singleton that encapsulates the MQTT client and .
type MQTTManager struct {
	actionTopic        string
	stateTopic         string
	autoDiscoveryTopic string
	listenerId         uuid.UUID
	mqttCfg            autopaho.ClientConfig
	connectionManager  *autopaho.ConnectionManager
}

// GetMQTTService returns the one and only MQTTService instance.
func GetMQTTService() *MQTTManager {
	once.Do(func() {
		instance = newMQTTService()
	})
	return instance
}

// Creates a new MQTTService object.
func newMQTTService() *MQTTManager {
	u, err := url.Parse(config.GetMQTTURL())
	if err != nil {
		panic(err)
	}
	mqttService := &MQTTManager{
		actionTopic:        fmt.Sprintf("%s/cover/%s/action", config.GetMQTTDiscoveryPrefix(), config.GetMQTTObjectID()),
		stateTopic:         fmt.Sprintf("%s/cover/%s/state", config.GetMQTTDiscoveryPrefix(), config.GetMQTTObjectID()),
		autoDiscoveryTopic: fmt.Sprintf("%s/cover/%s/config", config.GetMQTTDiscoveryPrefix(), config.GetMQTTObjectID()),
	}
	mqttCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		CleanStartOnInitialConnection: false,
		KeepAlive:                     30,
		SessionExpiryInterval:         0,
		OnConnectionUp:                mqttService.connectHandler,
		OnConnectError:                mqttService.connectErrorHandler,
		ClientConfig: paho.ClientConfig{
			ClientID:           config.GetMQTTClientID(),
			OnPublishReceived:  []func(paho.PublishReceived) (bool, error){mqttService.publishHandler},
			OnClientError:      mqttService.clientErrorHandler,
			OnServerDisconnect: mqttService.disconnectHandler,
		},
	}
	if config.GetMQTTUsername() != "" {
		mqttCfg.ConnectUsername = config.GetMQTTUsername()
		mqttCfg.ConnectPassword = []byte(config.GetMQTTPassword())
	}
	mqttService.mqttCfg = mqttCfg
	return mqttService
}

func (s *MQTTManager) Connect(ctx context.Context) error {
	cm, err := autopaho.NewConnection(ctx, s.mqttCfg)
	if err != nil {
		return fmt.Errorf("failed to create connection manager: %v", err)
	}
	s.connectionManager = cm
	return nil
}

func (s *MQTTManager) connectHandler(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	log.Info().Msgf("connected to MQTT broker: %s", connAck.String())

	// Subscribe to the action topic.
	if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: s.actionTopic,
				QoS:   1,
			},
		},
	}); err != nil {
		log.Error().Msgf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
	}
	log.Info().Msgf("subscribed to MQTT topic: %s", s.actionTopic)

	s.registerStateListener()
	s.sendHomeAssistantAutodiscoveryPayload()
}

func (s *MQTTManager) connectErrorHandler(err error) {
	log.Error().Msgf("mqtt connection error: %v", err)
}

func (s *MQTTManager) publishHandler(pr paho.PublishReceived) (bool, error) {
	dc := controller.GetDoorControllerService()
	command := string(pr.Packet.Payload)
	switch command {
	case "open", "close", "stop", "toggle":
		dc.RequestToggle()
	case "state":
		dc.RequestState()
	default:
		log.Warn().Msgf("received unknown command: %s", command)
		return false, fmt.Errorf("unknown command: %s", command)
	}

	log.Trace().Msgf("received command '%s' to DoorControllerService", command)
	return true, nil
}

func (s *MQTTManager) clientErrorHandler(err error) {
	log.Error().Msgf("mqtt client error: %v", err)
}

func (s *MQTTManager) disconnectHandler(d *paho.Disconnect) {
	if d.Properties != nil {
		log.Info().Msgf("server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		log.Info().Msgf("server requested disconnect; reason code: %d\n", d.ReasonCode)
	}
	dc := controller.GetDoorControllerService()
	if s.listenerId != uuid.Nil {
		dc.RemoveStateListener(s.listenerId)
	}
}

func (s *MQTTManager) registerStateListener() {
	dc := controller.GetDoorControllerService()
	if s.listenerId != uuid.Nil {
		dc.RemoveStateListener(s.listenerId)
	}
	s.listenerId = dc.AddStateListener(func(state string) {
		message := &paho.Publish{
			Topic:   s.stateTopic,
			Payload: []byte(state),
			QoS:     1,
		}
		if _, err := s.connectionManager.Publish(context.Background(), message); err != nil {
			log.Error().Msgf("failed to publish state (%s): %v", state, err)
		} else {
			log.Trace().Msgf("published state '%s' to MQTT topic: %s", state, s.stateTopic)
		}
	})

	// Delay the initial state update to ensure pins are at least read once.
	for i := 0; !dc.Ready(); i++ {
		time.Sleep(100 * time.Millisecond)
		if i > 50 {
			log.Warn().Msg("initial state update delayed too long")
			break
		}
	}

	dc.RequestState()
	log.Info().Msgf("registered state listener for MQTT topic: %s", s.stateTopic)
}

func (s *MQTTManager) sendHomeAssistantAutodiscoveryPayload() {
	iconTemplate := "{% if is_state(\"binary_sensor.voordeur_contact\", \"on\") %}\n" +
		"mdi:garage-open-variant\n" +
		"{% elif is_state(\"binary_sensor.voordeur_slot\", \"off\") %}\n" +
		"mdi:garage-variant\n" +
		"{% else %}\n" +
		"mdi:garage-alert-variant\n" +
		"{% endif %}"

	// Define the autodiscovery payload
	payload := map[string]interface{}{
		"name":          "Garage Door",
		"command_topic": s.actionTopic,
		"state_topic":   s.stateTopic,
		"payload_open":  "open",
		"payload_close": "close",
		"payload_stop":  "stop",
		"state_open":    "open",
		"state_closed":  "closed",
		"unique_id":     config.GetMQTTObjectID(),
		"icon":          iconTemplate,
		"device": map[string]interface{}{
			"identifiers":  config.GetMQTTObjectID(),
			"name":         "Garage Door",
			"model":        "Generic Garage Door",
			"manufacturer": "n/a",
		},
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Error().Msgf("failed to marshal autodiscovery payload: %v", err)
		return
	}

	// Publish the autodiscovery payload
	message := &paho.Publish{
		Topic:   s.autoDiscoveryTopic,
		Payload: payloadBytes,
		QoS:     1,
	}
	if _, err := s.connectionManager.Publish(context.Background(), message); err != nil {
		log.Error().Msgf("failed to publish autodiscovery payload: %v", err)
	} else {
		log.Info().Msgf("published autodiscovery payload to MQTT topic: %s", s.autoDiscoveryTopic)
	}
}
