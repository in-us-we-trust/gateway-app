package domain

type UsersManagerDeviceResponse struct {
	Data Device `json:"data"`
}

type EventsResponse struct {
	Data []Event `json:"data"`
}

type UserResponse struct {
	Data User `json:"data"`
}
