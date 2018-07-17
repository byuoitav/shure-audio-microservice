package commands

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/byuoitav/common/status"
	"github.com/fatih/color"
)

const TRANSMITTER_OFF = 65535
const CALCULATING = 65534
const BATT_TYPE = "LION"
const TIME_INTERVAL = 10

func GetBattery(connection *net.TCPConn, message string) (status.Battery, error) {

	log.Printf("Sending message: %s...", message)

	connection.Write([]byte(message))

	reader := bufio.NewReader(connection)
	response, err := reader.ReadString('>')
	if err != nil {
		errorMessage := "Error getting response: " + err.Error()
		log.Printf(errorMessage)
		return status.Battery{}, errors.New(errorMessage)
	}

	connection.Close()

	color.Set(color.FgHiGreen)
	log.Printf("Response: %s", string(response))
	color.Unset()

	log.Printf("Parsing device response...")
	re := regexp.MustCompile("[\\d]{3,5}")
	value := re.FindString(response)

	log.Printf("Device response: %s", value)

	timeRemaining, err := strconv.Atoi(value)
	if err != nil {
		errorMessage := "Could not parse time string: " + err.Error()
		log.Printf(errorMessage)
		return status.Battery{}, errors.New(errorMessage)
	}

	if timeRemaining == TRANSMITTER_OFF {
		errorMessage := "Transmitter deactivated."
		log.Printf(errorMessage)
		return status.Battery{}, errors.New(errorMessage)
	} else if timeRemaining == CALCULATING {
		errorMessage := "Currently calculating battery level"
		log.Printf(errorMessage)
		return status.Battery{}, errors.New(errorMessage)
	}

	return status.Battery{
		Battery: timeRemaining,
	}, nil
}

func GetPower(connection *net.TCPConn, channel string) (status.Power, error) {

	log.Printf("Getting power state of %s...", channel)

	message := fmt.Sprintf("< GET %s TX_TYPE >", channel)
	connection.Write([]byte(message))

	reader := bufio.NewReader(connection)
	response, err := reader.ReadString('>')
	if err != nil {
		errorMessage := "Error getting response: " + err.Error()
		log.Printf(errorMessage)
		return status.Power{}, errors.New(errorMessage)
	}

	connection.Close()

	//validate response
	if !strings.Contains(response, "TX_TYPE") { //got wrong response
		msg := color.RedString("[server] Erroneous response detected. Expected response containing \"TX_TYPE\", recieved \"%s\"", response)
		log.Printf(msg)
		return status.Power{}, errors.New(msg)
	}

	if strings.Contains(response, "UNKN") {
		return status.Power{
			Power: "standby",
		}, nil
	} else {
		return status.Power{
			Power: "on",
		}, nil
	}

}
