package handlers

import (
	"net/http"

	"github.com/byuoitav/shure-audio-microservice/commands"
	"github.com/labstack/echo"
)

func Health(context echo.Context) error {

	const HEALTH = "The fleet has moved out of lightspeed and we're preparing to - augh!"

	return context.JSON(http.StatusOK, HEALTH)

}

func Raw(context echo.Context) error {

	var command commands.RawCommand
	err := context.Bind(&command)
	if err != nil {
		return context.JSON(http.StatusBadRequest, err.Error())
	}

	response, err := commands.HandleRawCommand(command)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, err.Error())
	}

	return context.JSON(http.StatusOK, response)
}

func Battery(context echo.Context) error {

	channel := context.Param("channel")
	address := context.Param("address")
	format := context.Param("format")

	connection, err := Connect(address)
	if err != nil {
		return context.JSON(http.StatusBadRequest, "Could not connect to Shure device: "+err.Error())
	}

	message, err := GetMessage(format, channel)
	if err != nil {
		return context.JSON(http.StatusBadRequest, "Invalid format. Format must be \"time\" or \"percentage\"")
	}

	status, err := commands.GetBattery(connection, message)
	if err != nil {
		return context.JSON(http.StatusInternalServerError, "Error retrieving status: "+err.Error())
	}

	return context.JSON(http.StatusOK, status)
}

func Power(context echo.Context) error {

	connection, err := Connect(context.Param("address"))
	if err != nil {
		return context.JSON(http.StatusBadRequest, "Could not connect to Shure device: "+err.Error())
	}

	power, err := commands.GetPower(connection, context.Param("channel"))
	if err != nil {
		return context.JSON(http.StatusInternalServerError, "Error retrieving status: "+err.Error())
	}

	return context.JSON(http.StatusOK, power)
}
