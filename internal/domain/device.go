package domain

type Device struct {
	DeviceID  string `json:"deviceId"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

type LinkDeviceRequest struct {
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
}
