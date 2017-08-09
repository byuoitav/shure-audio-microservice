package handlers

import (
	"errors"
	"fmt"
	"log"
	"net"
)

const PORT = 2202

var MESSAGES = map[string]string{
	"percentage": "< GET %s BATT_CHARGE >",
	"time":       "< GET %s BATT_RUN_TIME >",
	"bars":       "< GET %s BATT_BARS >",
}

func ValidateChannel(conn net.Conn, channel string) error {

	return nil
}

func GetMessage(format, channel string) (string, error) {

	message := MESSAGES[format]

	if len(message) == 0 {
		return "", errors.New("Invalid format")
	}

	message = fmt.Sprintf(message, channel)

	return message, nil
}

func Connect(address string) (*net.TCPConn, error) {

	log.Printf("Connecting to %s...", address)

	ip := net.ParseIP(address)
	if ip == nil {
		return nil, errors.New("Invalid IP address")
	}

	addr := net.TCPAddr{
		IP:   ip,
		Port: PORT,
	}

	connection, err := net.DialTCP("tcp", nil, &addr)
	if err != nil {
		return nil, err
	}
	return connection, nil
}
