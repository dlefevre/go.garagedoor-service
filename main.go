package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/dlefevre/go.garagedoor-service/config"
	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/dlefevre/go.garagedoor-service/web"

	"github.com/rs/zerolog/log"
)

func main() {
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Info().Msg("Shutting down")
}
