package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dlefevre/go.garagedoor-service/config"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	instance *WebServiceImpl
	lock     sync.Mutex
)

type WebServiceImpl struct {
	echo    *echo.Echo
	apiKeys map[string]bool
}

func GetWebService() *WebServiceImpl {
	if instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if instance == nil {
			instance = newWebService()
		}
	}
	return instance
}

func newWebService() *WebServiceImpl {
	s := &WebServiceImpl{
		echo:    echo.New(),
		apiKeys: make(map[string]bool),
	}

	protected := s.echo.Group("")
	protected.Use(s.validateApiKey)

	s.echo.GET("/readyz", healthCheck)
	s.echo.GET("/healthz", healthCheck)
	protected.POST("/toggle", toggle)
	protected.GET("/state", state)

	return s
}

func (s *WebServiceImpl) Start() {
	address := fmt.Sprintf("%s:%d", config.GetBindHost(), config.GetBindPort())
	go func() {
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("%v", err)
		}
	}()
}

func (s *WebServiceImpl) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.echo.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("%v", err)
	}

}

func (s *WebServiceImpl) validateApiKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("x-api-key")
		if s.apiKeys[apiKey] {
			return next(c)
		}

		for _, digest := range config.GetApiKeys() {
			if err := bcrypt.CompareHashAndPassword([]byte(digest), []byte(apiKey)); err == nil {
				s.apiKeys[apiKey] = true
				return next(c)
			}
		}

		log.Warn().Msgf("Unauthorized request (forwarded ip: %v)", c.Request().Header.Get("x-forwarded-for"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			SimpleResponse: SimpleResponse{
				Result: "nok",
			},
			Message: "Unauthorized",
		})
	}
}
