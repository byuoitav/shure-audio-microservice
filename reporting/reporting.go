package reporting

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/byuoitav/common/db"
	"github.com/byuoitav/common/events"
	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
	eventhelper "github.com/byuoitav/shure-audio-microservice/event"
	"github.com/byuoitav/shure-audio-microservice/publishing"
	"github.com/byuoitav/shure-audio-microservice/state"
	"github.com/fatih/color"
)

const PORT = 2202

func Monitor(building, room string) {

	log.L.Infof("Starting mic reporting in building %s, room %s", building, room)

	//get Shure device
	log.L.Infof("Accessing shure device...")
	shure, err := db.GetDB().GetDevicesByRoomAndRole(fmt.Sprintf("%v-%v", building, room), "Receiver")
	for err != nil {
		log.L.Debugf("%s", color.HiRedString("[publisher] receiver not found: %s, retrying in 3s...", err.Error()))
		time.Sleep(5 * time.Second)
		shure, err = db.GetDB().GetDevicesByRoomAndRole(fmt.Sprintf("%v-%v", building, room), "Receiver")
	}

	if len(shure) == 0 {
		log.L.Debugf("%s", color.HiRedString("[publisher] no reciever detected in room. Aborting publisher..."))
		return
	}

	if len(shure) > 1 {
		msg := fmt.Sprintf("[error] detected %v recievers, expecting 1.", len(shure))
		log.L.Debugf("%s", color.HiRedString("[publisher] %s", msg))
		publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
		return
	}

	log.L.Infof("%s", color.HiBlueString("[reporting] connecting to device %s at address %s...", shure[0].Name, shure[0].Address))

	connection, err := net.DialTimeout("tcp", shure[0].Address+":2202", time.Second*3)
	if err != nil {
		errorMessage := fmt.Sprintf("[error] Could not connect to device: %s", err.Error())
		color.Set(color.FgHiYellow, color.Bold)
		log.L.Debugf(errorMessage)
		color.Unset()
		publishing.ReportError(errorMessage, shure[0].Name, building, room)
		return
	}

	reader := bufio.NewReader(connection)
	log.L.Infof("%s", color.HiGreenString("[reporting] successfully connected to device %s", shure[0].Name))
	log.L.Infof("%s", color.HiBlueString("[reporting] listening for events..."))

	for {

		data, err := reader.ReadString('>')
		if err != nil {
			msg := fmt.Sprintf("problem reading receiver string: %s", err.Error())
			publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
			continue
		}
		log.L.Debugf("%s", color.HiGreenString("[reporting] read string: %s", data))

		eventInfo, err := GetEventInfo(data)
		if err != nil {
			msg := fmt.Sprintf("problem reading receiver string: %s", err.Error())
			publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
		} else if eventInfo.Device == "" {
			continue
		}

		err = publishing.PublishEvent(false, eventInfo, building, room)
		if err != nil {
			msg := fmt.Sprintf("failed to publish event: %s", err.Error())
			publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
		}

	}

}

func GetEventInfo(data string) (events.EventInfo, error) {

	//identify device name
	re := regexp.MustCompile("REP [\\d]")
	channel := re.FindString(data)

	if len(channel) == 0 {
		msg := "no data"
		log.L.Debugf("%s", color.HiYellowString("[reporting] %s", msg))
		return events.EventInfo{}, nil
	}

	deviceName := fmt.Sprintf("MIC%s", channel[len(channel)-1:])

	log.L.Debugf("[resporting] device %s reporting", deviceName)
	data = re.ReplaceAllString(data, "")

	eventInfo := events.EventInfo{
		Device: deviceName,
	}

	//identify event type: interference, power, battery
	E, er := GetEventType(data)
	if er != nil {
		return events.EventInfo{}, nil
	}

	err := E.FillEventInfo(data, &eventInfo)
	if strings.EqualFold(eventInfo.EventInfoValue, "ignored") || len(eventInfo.EventInfoKey) == 0 {
		return events.EventInfo{}, nil
	}

	if err != nil {
		return eventInfo, err
	} else if eventInfo.EventInfoValue == eventhelper.FLAG {
		return eventInfo, nil
	}

	return eventInfo, nil
}

func GetEventType(data string) (eventhelper.Context, *nerr.E) {

	if strings.Contains(data, state.Interference.String()) {
		return eventhelper.Context{E: eventhelper.Interference{}}, nil
	} else if strings.Contains(data, state.Power.String()) {
		return eventhelper.Context{E: eventhelper.Power{}}, nil
	} else if strings.Contains(data, state.Battery.String()) {
		return eventhelper.Context{E: eventhelper.Battery{}}, nil
	} else {
		return eventhelper.Context{}, nerr.Create("Couldn't generate event type", "invalid")
	}
}
