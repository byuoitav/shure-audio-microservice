package commands

import (
	"log"
	"net"
)

type RawCommand struct {
	Address string `json:"address"`
	Message string `json:"message"`
	Port    string `json:"port"`
}

func HandleRawCommand(raw RawCommand) (string, error) {

	//build address
	address := raw.Address + ":" + raw.Port
	log.Printf("Address: %s", address)

	log.Printf("Connecting to device...")
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}

	defer connection.Close()

	log.Printf("Writing to connection...")
	connection.Write([]byte(raw.Message))

	buffer := make([]byte, 32)

	_, err = connection.Read(buffer)
	if err != nil {
		return "", err
	}

	return string(buffer), nil

}
