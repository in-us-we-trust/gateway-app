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

type UpdateDeviceRequest struct {
	City               *string `json:"city,omitempty"`
	DetectionThreshold *int    `json:"detectionThreshold,omitempty"`
	Location           *string `json:"location,omitempty"`
}

type DetectiveDevice struct {
	ID                 string `json:"id"`
	SerialNumber       string `json:"serialNumber"`
	City               string `json:"city"`
	LastSeen           string `json:"lastSeen"`
	LastBoot           string `json:"lastBoot"`
	Status             string `json:"status"`
	Uptime             int    `json:"uptime"`
	DetectionThreshold int    `json:"detectionThreshold"`
	Location           string `json:"location"`
	FirmwareVersion    string `json:"firmwareVersion"`
}

type DeviceResponseDto struct {
	DeviceID           string `json:"deviceId"`
	Name               string `json:"name"`
	CreatedAt          string `json:"createdAt"`
	SerialNumber       string `json:"serialNumber"`
	City               string `json:"city"`
	LastSeen           string `json:"lastSeen"`
	LastBoot           string `json:"lastBoot"`
	Status             string `json:"status"`
	Uptime             int    `json:"uptime"`
	DetectionThreshold int    `json:"detectionThreshold"`
	Location           string `json:"location"`
	FirmwareVersion    string `json:"firmwareVersion"`
}

func ToDeviceResponseDto(device *Device, dd *DetectiveDevice) *DeviceResponseDto {
	deviceResponse := &DeviceResponseDto{
		DeviceID:  device.DeviceID,
		Name:      device.Name,
		CreatedAt: device.CreatedAt,
	}
	if dd != nil {
		deviceResponse.SerialNumber = dd.SerialNumber
		deviceResponse.City = dd.City
		deviceResponse.LastSeen = dd.LastSeen
		deviceResponse.LastBoot = dd.LastBoot
		deviceResponse.Status = dd.Status
		deviceResponse.Uptime = dd.Uptime
		deviceResponse.DetectionThreshold = dd.DetectionThreshold
		deviceResponse.Location = dd.Location
		deviceResponse.FirmwareVersion = dd.FirmwareVersion
	}
	return deviceResponse
}
