package mqtt

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/dlefevre/go.garagedoor-service/controller"
	mochi_mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/hooks/debug"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	done   = make(chan bool, 1)
	server = mochi_mqtt.New(&mochi_mqtt.Options{
		InlineClient: true,
	})
	state string = "unknown"
)

func init() {
	os.Setenv("GARAGESERVICE_CONFIG_PATH", "..")
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	// Allow all connections.
	_ = server.AddHook(new(auth.AllowHook), &auth.Options{
		Ledger: &auth.Ledger{
			Auth: auth.AuthRules{ // Auth disallows all by default
				{Username: "test", Password: "test", Allow: true},
				{Remote: "127.0.0.1:*", Allow: true},
				{Remote: "localhost:*", Allow: true},
			},
			ACL: auth.ACLRules{ // ACL allows all by default
				{Remote: "127.0.0.1:*"},
				{
					Username: "test", Filters: auth.Filters{
						"homeassistent/#":   auth.ReadWrite,
						"homeassistent/+/+": auth.ReadWrite,
					},
				},
			},
		},
	})
	_ = server.AddHook(new(debug.Hook), nil)

	// Create a TCP listener on a standard port.
	tcp := listeners.NewTCP(listeners.Config{ID: "t1", Address: ":1883"})
	if err := server.AddListener(tcp); err != nil {
		log.Fatal().Msgf("Could not add Listener: %v", err)
	}

	server.Subscribe("homeassistant/+/+/+", 0, func(cl *mochi_mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		log.Info().Msgf("Received message: %s", pk.Payload)
		state = string(pk.Payload)
	})
}

func startBroker() {
	go func() {
		err := server.Serve()
		if err != nil {
			log.Fatal().Msgf("Could not start server: %v", err)
		}
	}()

	// Run server until interrupted
	<-done
}

func TestSingleton(t *testing.T) {
	mqttService := GetMQTTService()
	if mqttService == nil {
		t.Fatalf("Expected MQTTService to be non-nil")
	}
	if GetMQTTService() != mqttService {
		t.Fatalf("Expected GetMQTTService to return the same instance")
	}
}

func TestConnect(t *testing.T) {
	mqttService := GetMQTTService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := mqttService.Connect(ctx); err != nil {
		t.Fatalf("Error connecting to MQTT broker: %v", err)
	}
}

func TestPublish(t *testing.T) {
	go startBroker()
	time.Sleep(1 * time.Second)

	mqttService := GetMQTTService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dc := controller.GetDoorControllerService()
	dc.Start()

	if err := mqttService.Connect(ctx); err != nil {
		t.Fatalf("Error connecting to MQTT broker: %v", err)
	}
	time.Sleep(1 * time.Second)
	if state != "closed" {
		t.Fatalf("Expected state to be closed, got %s", state)
	}

	log.Info().Msg("Publishing message to action topic")
	if err := server.Publish("homeassistant/cover/garage_door/action", []byte("open"), false, 1); err != nil {
		t.Fatalf("Error publishing message: %v", err)
	}

	time.Sleep(1 * time.Second)
	if state != "open" {
		t.Fatalf("Expected state to be open, got %s", state)
	}
	done <- true
}
