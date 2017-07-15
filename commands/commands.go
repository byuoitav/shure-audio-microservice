package commands

import (
	"io/ioutil"
	"log"
	"net"

	"github.com/byuoitav/av-api/status"
)

func GetBatteryLevel(connection *net.TCPConn, address, channel string) (status.Battery, error) {

	log.Printf("Getting battery level of mic: %s", channel)

	log.Printf("Sending request...")
	message := []byte("GET " + channel + " BATT_RUN_TIME")
	connection.Write(message)

	response, err := ioutil.ReadAll(connection)
	if err != nil {
		return status.Battery{}, err
	}

	connection.Close()

	log.Printf("Response: %s", string(response))
	return status.Battery{}, nil
}
