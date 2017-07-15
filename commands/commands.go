package commands

import (
	"bufio"
	"errors"
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

func GetBatteryLevel(connection *net.TCPConn, address, channel string) (status.Battery, error) {

	log.Printf("Getting battery level of mic: %s", channel)

	log.Printf("Sending request...")
	message := []byte("< GET " + channel + " BATT_RUN_TIME >")
	connection.Write(message)

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
	re := regexp.MustCompile("[\\d]{5}")
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
