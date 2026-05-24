package domain

type EventType string

const (
	EventTypeBell   EventType = "bell"
	EventTypeMotion EventType = "motion"
)

type EventRequest struct {
	Type  EventType `json:"type"  binding:"required,oneof=bell motion"`
	Image string    `json:"image" binding:"required"`
}

type EventData struct {
	IsIntruder bool `json:"is_intruder"`
}

type EventResponse struct {
	Data EventData `json:"data"`
}

type Event struct {
	ID        string `json:"id"`
	DeviceID  string `json:"deviceId"`
	Timestamp string `json:"timestamp"`
	Intruder  bool   `json:"intruder"`
	ImageURL  string `json:"imageUrl"`
	EventType string `json:"eventType"`
}
