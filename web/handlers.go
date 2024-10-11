package web

import (
	"net/http"

	"github.com/dlefevre/go.garagedoor-service/controller"
	"github.com/labstack/echo/v4"
)

type SimpleResponse struct {
	Result string `json:"result"`
}

type ErrorResponse struct {
	SimpleResponse
	Message string `json:"message"`
}

type StateResponse struct {
	SimpleResponse
	State string `json:"state"`
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, SimpleResponse{
		Result: "OK",
	})
}

func toggle(c echo.Context) error {
	dc := controller.GetDoorControllerService()
	dc.RequestToggle()
	return c.JSON(http.StatusOK, SimpleResponse{
		Result: "ok",
	})
}

func state(c echo.Context) error {
	dc := controller.GetDoorControllerService()
	return c.JSON(http.StatusOK, StateResponse{
		SimpleResponse: SimpleResponse{
			Result: "ok",
		},
		State: dc.GetStateStr(),
	})
}
