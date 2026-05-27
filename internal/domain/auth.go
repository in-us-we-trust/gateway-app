package domain

type AuthRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
