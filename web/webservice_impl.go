package web

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dlefevre/go.garagedoor-service/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

var (
	instance *WebServiceImpl
	once     sync.Once
)

type WebServiceImpl struct {
	echo    *echo.Echo
	apiKeys map[string]bool
}

// Get one and only WebServiceImpl instance.
func GetWebService() *WebServiceImpl {
	once.Do(func() {
		instance = newWebService()
	})
	return instance
}

// Creates a new WebServiceImpl object.
func newWebService() *WebServiceImpl {
	return &WebServiceImpl{
		echo:    nil,
		apiKeys: make(map[string]bool),
	}
}

// Configure the Echo web server.
func (s *WebServiceImpl) setUpEcho() {
	s.echo = echo.New()

	s.echo.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:        true,
		LogStatus:     true,
		LogValuesFunc: logger,
	}))

	s.echo.GET("/readyz", healthCheck)
	s.echo.GET("/healthz", healthCheck)

	protected := s.echo.Group("")
	protected.Use(s.validateApiKey)

	protected.POST("/toggle", toggle)
	protected.GET("/state", state)
	protected.GET("/ws", ws)
}

// Start the web server.
func (s *WebServiceImpl) Start() {
	s.setUpEcho()
	address := fmt.Sprintf("%s:%d", config.GetBindHost(), config.GetBindPort())
	go func() {
		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("%v", err)
		}
	}()
}

// Stop the web server.
func (s *WebServiceImpl) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.echo.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("%v", err)
	}
	s.echo = nil

}

// Middleware handler to validate the API key. The API key is first matched against an internal
// cache of valid keys, then against the list of keys in the configuration file.
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

		log.Warn().Msgf("Unauthorized request to %v (forwarded ip: %v)",
			c.Request().RequestURI,
			c.Request().Header.Get("x-forwarded-for"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			SimpleResponse: SimpleResponse{
				Result: "nok",
			},
			Message: "Unauthorized",
		})
	}
}

// Middleware handler to log requests.
func logger(c echo.Context, v middleware.RequestLoggerValues) error {
	log.Info().
		Str("URI", v.URI).
		Int("status", v.Status).
		Msg("request")

	return nil
}
