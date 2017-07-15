package commands

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"

	"github.com/byuoitav/av-api/status"
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
	log.Printf("Response: %s", string(response))

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

func GetPower(connection *net.TCPConn, channel string) (status.PowerStatus, error) {

	log.Printf("Getting power state of %s...", channel)

	message := fmt.Sprintf("< GET %s BATT_RUN_TIME >", channel)
	connection.Write([]byte(message))

	reader := bufio.NewReader(connection)
	response, err := reader.ReadString('>')
	if err != nil {
		errorMessage := "Error getting response: " + err.Error()
		log.Printf(errorMessage)
		return status.PowerStatus{}, errors.New(errorMessage)
	}

	connection.Close()
	log.Printf("Response: %s", response)

	re := regexp.MustCompile("[\\d]{5}")
	value := re.FindString(response)

	log.Printf("Value: %s", value)

	if value == "65535" {
		return status.PowerStatus{
			Power: "standby",
		}, nil
	} else {
		return status.PowerStatus{
			Power: "on",
		}, nil
	}

}
