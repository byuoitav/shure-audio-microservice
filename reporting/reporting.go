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

	const MIC = "MIC"

	//identify device name
	//TODO commit to naming scheme???
	re := regexp.MustCompile("[\\d]{1}")
	channel := re.FindString(data)
	deviceName := MIC + channel

	eventInfo := ei.EventInfo{
		Device: deviceName,
	}

	//identify event type: interference, power on, power off
	switch GetEventType(data) {

	case state.Interference:

		err := ProcessInterference(&eventInfo)
		if err != nil {
			return nil, err
		}

	case state.Power:

		err := ProcessPower(&eventInfo)
		if err != nil {
			return nil, err
		}

	case state.Battery:

		err := ProcessBattery(&eventInfo)
		if err != nil {
			return nil, err
		}

	default:
		return nil, nil
	}

	return &eventInfo, nil
}

func PublishEvent(eventInfo *ei.EventInfo, building, room string) error {

	if eventInfo == nil {
		return nil
	}

	event := ei.Event{
		Hostname:  building + "-" + room + "-" + eventInfo.Device,
		Timestamp: time.Now().Format(time.RFC3339)
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

func GetEventType(data string) state.State {

	if strings.Contains(data, state.Interference.String()) {
		return state.Interference
	} else if strings.Contains(data, state.Power.String()) {
		return state.Power
	} else if strings.Contains(data, state.Battery.String()) {
		return state.Battery
	} else {
		return state.Unknown
	}
}

//fills EventInfo.EventInfoKey, EventInfo.EventInfoValue, EventInfo.Type, and EventInfo.EventCause
func ProcessInterference(eventInfo *ei.EventInfo) error {
	return nil
}

//fills EventInfo.EventInfoKey, EventInfo.EventInfoValue, EventInfo.Type, and EventInfo.EventCause
func ProcessPower(eventInfo *ei.EventInfo) error {
	return nil
}

//fills EventInfo.EventInfoKey, EventInfo.EventInfoValue, EventInfo.Type, and EventInfo.EventCause
func ProcessBattery(eventInfo *ei.EventInfo) error {
	return nil
}
