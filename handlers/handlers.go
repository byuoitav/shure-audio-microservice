package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

func Health(context echo.Context) error {

	message := "The fleet has moved out of lightspeed and we're preparing to - augh!"

	return context.JSON(http.StatusOK, message)

}
