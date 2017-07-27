package publishing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	sb "github.com/byuoitav/event-router-microservice/subscription"
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
			var request sb.SubscribeRequest
			request.Address = "localhost:" + PORT
			body, err := json.Marshal(request)
			if err != nil {
				color.Set(color.FgRed)
				log.Printf("[error] %s", err.Error())
				color.Unset()
			}

			_, err = http.Post("http://localhost:6999/subscribe", "application/json", bytes.NewBuffer(body))
			for err != nil {
				_, err = http.Post("http://localhost:6999/subscribe", "application/json", bytes.NewBuffer(body))
				color.Set(color.FgRed)
				log.Printf("[error] The router hasn't subsribed to me yet. Retrying...")
				color.Unset()
				time.Sleep(SLEEP_INTERVAL * time.Second)
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
