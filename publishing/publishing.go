package publishing

import (
	"log"
	"os"
	"time"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

const CHAN_SIZE = 10
const SLEEP_INTERVAL = 3

var node *ei.EventNode

func Start() {
	log.Printf("Starting node...")
	node = ei.NewEventNode("Shure", os.Getenv("EVENT_ROUTER_ADDRESS"), []string{})
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

	log.Printf("Publishing event: %v", event)

	header := ""
	if isError {
		header = ei.APIError
	} else {
		header = ei.Metrics
	}

	log.Printf("header: %s", header)

	return node.PublishEvent(event, header)
}

func ReportError(err, device, building, room string) error {

	log.Printf("reporting error: %s", err)

	eventInfo := ei.EventInfo{
		EventInfoKey:   "Error String",
		EventInfoValue: err,
		Device:         device,
		Type:           ei.ERROR,
		EventCause:     ei.INTERNAL,
	}

	PublishEvent(true, &eventInfo, building, room)

	return nil
}
