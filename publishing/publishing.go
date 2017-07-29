package publishing

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/fatih/color"
	"github.com/xuther/go-message-router/common"
	pb "github.com/xuther/go-message-router/publisher"
)

const PORT = "7005"
const QUEUE_SIZE = 1000
const CHAN_SIZE = 10
const SLEEP_INTERVAL = 3

var Publisher pb.Publisher

func Start() {

	log.Printf("Starting publisher...")

	var err error
	Publisher, err = pb.NewPublisher(PORT, QUEUE_SIZE, CHAN_SIZE)
	if err != nil {
		errorMessage := fmt.Sprintf("[publisher] Unable to start publisher. Error: %s\n", err.Error())
		color.Set(color.FgRed)
		log.Fatalf(errorMessage)
	}

	go func() {
		log.Printf("Listening...")
		err = Publisher.Listen()
		if err != nil {
			errorMessage := fmt.Sprintf("[publisher] Unable to start publisher. Error: %s\n", err.Error())
			color.Set(color.FgRed)
			log.Fatalf(errorMessage)
		} else {
			color.Set(color.FgGreen)
			log.Printf("[publisher] Publisher started on port %s", PORT)
			color.Unset()
		}
	}()

	if len(os.Getenv("LOCAL_ENVIRONMENT")) > 0 {

		log.Printf("Local environment detected.")

		go func() {
			var request ei.ConnectionRequest
			request.PublisherAddr = "localhost:" + PORT
			err = ei.SendConnectionRequest("http://localhost:6999/subscribe", request, true)
			if err != nil {
				color.Set(color.FgRed)
				log.Fatalf("[error] Could not connect to event router microservice")
			}
		}()
	}
}

func PublishEvent(isError bool, eventInfo *ei.EventInfo, building, room string) error {

	if eventInfo == nil {
		return nil
	}

	event := ei.Event{
		Hostname:  building + "-" + room + "-" + eventInfo.Device,
		Timestamp: time.Now().Format(time.RFC3339),
		Event:     *eventInfo,
		Building:  building,
		Room:      room,
	}

	//get local environment
	localEnvironment := os.Getenv("LOCAL_ENVIRONMENT")
	if len(localEnvironment) > 0 {
		event.LocalEnvironment = true
	} else {
		event.LocalEnvironment = false
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	log.Printf("Publishing event: %s", body)

	header := [24]byte{}
	copy(header[:], ei.Metrics)

	log.Printf("header: %s", header)
	err = Publisher.Write(common.Message{MessageHeader: header, MessageBody: body})
	return err
}

func ReportError(err string) error {

	log.Printf("reporting error: %s", err)

	return nil
}
