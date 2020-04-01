package types

type Event struct {
	EventType        string `json:"event-type"`
	EventTypeVersion string `json:"event-type-version"`
	EventId          string `json:"event-id"`
	EventTime        string `json:"event-time"`
	Data             string `json:"data"`
}
