package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dlefevre/go.garagedoor-service/config"
	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/dlefevre/go.garagedoor-service/mqtt"
	"github.com/dlefevre/go.garagedoor-service/web"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().Msg("Verifying configuration")
	config.Verify()

	log.Info().Msg("Starting Door Controller Service")
	dc := controller.GetDoorControllerService()
	dc.Start()
	defer dc.Stop()

	log.Info().Msg("Starting Web Service")
	ws := web.GetWebService()
	ws.Start()
	defer ws.Stop()

	if config.GetMQTTEnabled() {
		log.Info().Msg("Setting connection to MQTT Broker")
		ms := mqtt.GetMQTTService()
		if err := ms.Connect(ctx); err != nil {
			log.Fatal().Msgf("Error connecting to MQTT broker: %v", err)
		}
	}

	<-ctx.Done()
	log.Info().Msg("Shutting down")
}
