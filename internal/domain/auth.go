package domain

import "encoding/json"

type AuthRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthLoginRequest struct {
	Email    string          `json:"email"`
	Password string          `json:"password"`
	User     json.RawMessage `json:"user,omitempty"`
}

type LoginData struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type LoginResponse struct {
	Data LoginData `json:"data"`
}
