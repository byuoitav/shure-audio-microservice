package reporting

import (
	"bufio"
	"fmt"
	"log"
	"net"
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
	if err != nil {
		color.Set(color.FgRed)
		log.Printf("Could not get Shure device: %s", err.Error())
		color.Unset()
	}

	if len(shure) != 1 {
		errorMessage := fmt.Sprintf("[error] detected %v recievers, expecting 1.", len(shure))
		color.Set(color.FgRed)
		log.Printf(errorMessage)
		color.Unset()
		publishing.ReportError(errorMessage)
		return
	}

	log.Printf("Connecting to device %s at address %s...", shure[0].Name, shure[0].Address)
	connection, err := net.DialTimeout("tcp", shure[0].Address+":2202", time.Second*3)
	if err != nil {
		errorMessage := fmt.Sprintf("[error] Could not connect to device: %s", err.Error())
		color.Set(color.FgRed)
		log.Printf(errorMessage)
		color.Unset()
		publishing.ReportError(errorMessage)
		return
	}

	reader := bufio.NewReader(connection)
	color.Set(color.FgGreen)
	log.Printf("Successfully connected to device %s", shure[0].Name)
	color.Unset()

	for {

		data, err := reader.ReadString('>')
		if err != nil {
			errorMessage := "Error reading Shure string: " + err.Error()
			publishing.ReportError(errorMessage)
		}

		color.Set(color.FgGreen)
		log.Printf("Read string: %s", data)
		color.Unset()

		eventInfo, err := ParseString(data)
		if err != nil {
			errorMessage := "Error parsing Shure string: " + err.Error()
			publishing.ReportError(errorMessage)
		}

		err = publishing.PublishEvent(false, eventInfo, building, room)
		if err != nil {
			errorMessage := "Could not publish event: " + err.Error()
			publishing.ReportError(errorMessage)
		}

	}

}

func ParseString(data string) (*ei.EventInfo, error) {

	//identify device name
	re := regexp.MustCompile("[\\d]{1}")
	channel := re.FindString(data)
	deviceName := "MIC" + channel

	log.Printf("Device %s reporting", deviceName)

	eventInfo := ei.EventInfo{
		Device: deviceName,
	}

	//identify event type: interference, power, battery
	event := GetEventType(data)
	if event == nil {
		return nil, nil
	}

	err := event.FillEventInfo(data, &eventInfo)
	if err != nil {
		return nil, err
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
