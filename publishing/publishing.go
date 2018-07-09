package publishing

import (
	"os"
	"strings"
	"time"

	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
)

const CHAN_SIZE = 10
const SLEEP_INTERVAL = 3

var node *events.EventNode

func Start() {
	log.L.Infof("Starting node...")
	node = events.NewEventNode("Shure", os.Getenv("EVENT_ROUTER_ADDRESS"), []string{})
}

func PublishEvent(isError bool, eventInfo events.EventInfo, building, room string) error {
	if eventInfo.Device == "" || strings.EqualFold(eventInfo.EventInfoKey, "ignored") {
		return nil
	}

	event := events.Event{
		Hostname:  building + "-" + room + "-" + eventInfo.Device,
		Timestamp: time.Now().Format(time.RFC3339),
		Event:     eventInfo,
		Building:  building,
		Room:      room,
	}
	log.L.Debugf("Publishing event %+v", event)

	//get local environment
	localEnvironment := os.Getenv("LOCAL_ENVIRONMENT")
	if len(localEnvironment) > 0 {
		event.LocalEnvironment = true
	} else {
		event.LocalEnvironment = false
	}

	header := ""
	if isError {
		header = events.APIError
	} else {
		header = events.Metrics
	}

	log.L.Debugf("header: %s", header)

	return node.PublishEvent(header, event)
}

func ReportError(err, device, building, room string) error {

	log.L.Debugf("reporting error: %s", err)

	eventInfo := events.EventInfo{
		EventInfoKey:   "Error String",
		EventInfoValue: err,
		Device:         device,
		Type:           events.ERROR,
		EventCause:     events.INTERNAL,
	}

	PublishEvent(true, eventInfo, building, room)

	return nil
}
