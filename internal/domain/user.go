package domain

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	CreatedAt string `json:"createdAt"`
}

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type AppDevice struct {
	DeviceName string `json:"deviceName"`
	Platform   string `json:"platform"`
	PushToken  string `json:"pushToken"`
	LastSeenAt string `json:"lastSeenAt"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type UserDetail struct {
	ID         string      `json:"id"`
	Email      string      `json:"email"`
	FirstName  string      `json:"firstName"`
	LastName   string      `json:"lastName"`
	CreatedAt  string      `json:"createdAt"`
	UpdatedAt  string      `json:"updatedAt"`
	AppDevices []AppDevice `json:"appDevices"`
}
