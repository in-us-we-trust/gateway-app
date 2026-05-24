package domain

type DeviceResponse struct {
	Data Device `json:"data"`
}

type EventsResponse struct {
	Data []Event `json:"data"`
}
