package reporting

import (
	"bufio"
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
)

const PORT = 2202

func Monitor(building, room string) {

	log.Printf("Starting mic reporting in building %s, room %s", building, room)

	//get Shure device
	shure, err := dbo.GetDevicesByBuildingAndRoomAndRole(building, room, "Receiver")
	if err != nil {
		errorMessage := "Could not get Receiver configuration: " + err.Error()
		log.Printf(errorMessage)
		publishing.ReportError(errorMessage)
	}

	if len(shure) != 1 {
		errorMessage := "Invalid reciever configuration detected"
		log.Printf(errorMessage)
		publishing.ReportError(errorMessage)
	}

	ip := net.ParseIP(shure[0].Address)

	address := net.TCPAddr{
		IP:   ip,
		Port: PORT,
	}

	connection, err := net.DialTCP("tcp", nil, &address)
	if err != nil {
		errorMessage := "Could not connect to device: " + err.Error()
		log.Printf(errorMessage)
		publishing.ReportError(errorMessage)
	}

	reader := bufio.NewReader(connection)

	for {

		data, err := reader.ReadString('>')
		if err != nil {
			errorMessage := "Error reading Shure string: " + err.Error()
			publishing.ReportError(errorMessage)
		}

		eventInfo, err := ParseString(data)
		if err != nil {
			errorMessage := "Error parsing Shure string: " + err.Error()
			publishing.ReportError(errorMessage)
		}

		err = PublishEvent(eventInfo, building, room)
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

	err := event.FillEventInfo(data, &eventInfo)
	if err != nil {
		return nil, err
	}

	return &eventInfo, nil
}

func PublishEvent(eventInfo *ei.EventInfo, building, room string) error {

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

	return nil
}

func GetEventType(data string) event.Context {

	if strings.Contains(data, state.Interference.String()) {
		return event.Context{E: event.Interference{}}
	} else if strings.Contains(data, state.Power.String()) {
		return event.Context{E: event.Power{}}
	} else if strings.Contains(data, state.Battery.String()) {
		return event.Context{E: event.Battery{}}
	} else {
		return event.Context{}
	}
}
