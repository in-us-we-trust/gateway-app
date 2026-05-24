package domain

type JWTClaims struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id,omitempty"`
}
