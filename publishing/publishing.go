package publishing

import (
	"os"
	"strings"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	"github.com/byuoitav/common/v2/events"
)

const CHAN_SIZE = 10
const SLEEP_INTERVAL = 3

var m *messenger.Messenger

func Start() {
	log.L.Infof("Starting messenger...")
	var err *nerr.E
	m, err = messenger.BuildMessenger(os.Getenv("HUB_ADDRESS"), base.Messenger, 1000)
	if err != nil {

	}
}

func PublishEvent(isError bool, event events.Event, building, room string) error {
	if event.TargetDevice.DeviceID == "" || strings.EqualFold(event.Key, "ignored") {
		return nil
	}
	event.GeneratingSystem = event.TargetDevice.DeviceID
	event.Timestamp = time.Now()
	event.AffectedRoom = events.GenerateBasicRoomInfo(building + "-" + room)

	log.L.Debugf("Publishing event %+v", event)

	//get room system
	roomSystem := os.Getenv("ROOM_SYSTEM")
	if len(roomSystem) > 0 {
		event.AddToTags(roomSystem)
	}

	header := ""
	if isError {
		event.AddToTags(events.Error)
		header = events.Error
	} else {
		event.AddToTags(events.Error)
		header = events.Metrics
	}

	log.L.Debugf("header: %s", header)

	m.SendEvent(event)
	return nil
}

func ReportError(err, device, building, room string) error {

	log.L.Debugf("reporting error: %s", err)

	event := events.Event{
		Key:          "Error String",
		Value:        err,
		TargetDevice: events.GenerateBasicDeviceInfo(device),
	}
	event.AddToTags(events.Error, events.Internal)
	PublishEvent(true, event, building, room)

	return nil
}
