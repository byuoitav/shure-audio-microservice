package reporting

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/byuoitav/av-api/dbo"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/shure-audio-microservice/event"
	"github.com/byuoitav/shure-audio-microservice/publishing"
	"github.com/byuoitav/shure-audio-microservice/state"
	"github.com/fatih/color"
)

const PORT = 2202

func Monitor(building, room string) {

	log.Printf("Starting mic reporting in building %s, room %s", building, room)

	//get Shure device
	log.Printf("Accessing shure device...")
	shure, err := dbo.GetDevicesByBuildingAndRoomAndRole(building, room, "Receiver")
	for err != nil {
		log.Printf("%s", color.HiRedString("[publisher] receiver not found: %s, retrying in 3s...", err.Error()))
		time.Sleep(3 * time.Second)
		shure, err = dbo.GetDevicesByBuildingAndRoomAndRole(building, room, "Receiver")
	}

	if len(shure) == 0 {
		log.Printf("%s", color.HiRedString("[publisher] no reciever detected in room. Aborting publisher..."))
		return
	}

	if len(shure) > 1 {
		msg := fmt.Sprintf("[error] detected %v recievers, expecting 1.", len(shure))
		log.Printf("%s", color.HiRedString("[publisher] %s", msg))
		publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
		return
	}

	log.Printf("%s", color.HiBlueString("[reporting] connecting to device %s at address %s...", shure[0].Name, shure[0].Address))

	connection, err := net.DialTimeout("tcp", shure[0].Address+":2202", time.Second*3)
	if err != nil {
		errorMessage := fmt.Sprintf("[error] Could not connect to device: %s", err.Error())
		color.Set(color.FgHiYellow, color.Bold)
		log.Printf(errorMessage)
		color.Unset()
		publishing.ReportError(errorMessage, shure[0].Name, building, room)
		return
	}

	reader := bufio.NewReader(connection)
	log.Printf("%s", color.HiGreenString("[reporting] successfully connected to device %s", shure[0].Name))
	log.Printf("%s", color.HiBlueString("[reporting] listening for events..."))

	for {

		data, err := reader.ReadString('>')
		if err != nil {
			msg := fmt.Sprintf("problem reading receiver string: %s", err.Error())
			publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
			continue
		}

		log.Printf("%s", color.HiGreenString("[reporting] read string: %s", data))

		eventInfo, err := GetEventInfo(data)
		if err != nil {
			msg := fmt.Sprintf("problem reading receiver string: %s", err.Error())
			publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
		} else if eventInfo == nil {
			continue
		}

		err = publishing.PublishEvent(false, eventInfo, building, room)
		if err != nil {
			msg := fmt.Sprintf("failed to publish event: %s", err.Error())
			publishing.ReportError(msg, os.Getenv("PI_HOSTNAME"), building, room)
		}

	}

}

func GetEventInfo(data string) (*ei.EventInfo, error) {

	//identify device name
	re := regexp.MustCompile("REP [\\d]")
	channel := re.FindString(data)

	if len(channel) == 0 {
		msg := "no data"
		log.Printf("%s", color.HiYellowString("[reporting] %s", msg))
		return nil, nil
	}

	deviceName := fmt.Sprintf("MIC%s", channel[len(channel)-1:])

	log.Printf("[resporting] device %s reporting", deviceName)
	data = re.ReplaceAllString(data, "")

	eventInfo := ei.EventInfo{
		Device: deviceName,
	}

	//identify event type: interference, power, battery
	Event := GetEventType(data)
	if Event == nil {
		return nil, nil
	}

	err := Event.FillEventInfo(data, &eventInfo)
	if err != nil {
		return nil, err
	} else if eventInfo.EventInfoValue == event.FLAG {
		return nil, nil
	}

	return &eventInfo, nil
}

func GetEventType(data string) *event.Context {

	if strings.Contains(data, state.Interference.String()) {
		return &event.Context{E: event.Interference{}}
	} else if strings.Contains(data, state.Power.String()) {
		return &event.Context{E: event.Power{}}
	} else if strings.Contains(data, state.Battery.String()) {
		return &event.Context{E: event.Battery{}}
	} else {
		return nil
	}
}
