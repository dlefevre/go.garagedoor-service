package mqtt

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/dlefevre/go.garagedoor-service/config"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
)

var (
	instance *MQTTService
	once     sync.Once
)

// MQTTService is a singleton that encapsulates the MQTT client and .
type MQTTService struct {
}

// GetMQTTService returns the one and only MQTTService instance.
func GetMQTTService() *MQTTService {
	once.Do(func() {
		instance = newMQTTService()
	})
	return instance
}

// Creates a new MQTTService object.
func newMQTTService() *MQTTService {
	u, err := url.Parse(config.GetMQTTURL())
	if err != nil {
		panic(err)
	}
	mqttService := &MQTTService{}
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
	return &MQTTService{}
}

func (s *MQTTService) connectHandler(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	fmt.Println("mqtt connection up")
	// Subscribing in the OnConnectionUp callback is recommended (ensures the subscription is reestablished if
	// the connection drops)
	if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: fmt.Sprintf("%s%s/command", config.GetMQTTTopicPrefix(), config.GetMQTTClientID()),
				QoS:   1,
			},
		},
	}); err != nil {
		log.Error().Msgf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
	}
	fmt.Println("mqtt subscription made")
}

func (s *MQTTService) connectErrorHandler(err error) {
	log.Error().Msgf("mqtt connection error: %v", err)
}

func (s *MQTTService) publishHandler(pr paho.PublishReceived) (bool, error) {
	fmt.Printf("received message on topic %s; body: %s (retain: %t)\n", pr.Packet.Topic, pr.Packet.Payload, pr.Packet.Retain)
	return true, nil
}

func (s *MQTTService) clientErrorHandler(err error) {
	log.Error().Msgf("mqtt client error: %v", err)
}

func (s *MQTTService) disconnectHandler(d *paho.Disconnect) {
	if d.Properties != nil {
		fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
	}
}
