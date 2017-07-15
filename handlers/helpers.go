package handlers

import (
	"errors"
	"log"
	"net"
)

const PORT = 2202

func ValidateChannel(conn net.Conn, channel string) error {

	return nil
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
