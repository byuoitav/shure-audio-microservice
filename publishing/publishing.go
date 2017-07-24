package publishing

import "log"

type MicEvent struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func Subscribe() {}

func PublishEvent(event MicEvent) error {

	log.Printf("Publishing event: %s, %s", event.Key, event.Value)
	return nil
}

func ReportError(err string) error {

	log.Printf("reporting error: %s", err)

	return nil
}
