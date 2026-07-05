package domain

type NotificationToken struct {
	AppDeviceId int    `json:"appDeviceId"`
	DeviceName  string `json:"deviceName"`
	Platform    string `json:"platform"`
	Token       string `json:"token"`
}

type Contact struct {
	ContactId          string              `json:"contactId"`
	UserId             string              `json:"userId"`
	Name               string              `json:"name"`
	LastName           string              `json:"lastName"`
	Email              string              `json:"email"`
	IsOwner            bool                `json:"isOwner"`
	NotificationTokens []NotificationToken `json:"notificationTokens"`
}

type CreateContactRequest struct {
	Name     string `json:"name" binding:"required"`
	LastName string `json:"lastName" binding:"required"`
	Email    string `json:"email" binding:"required"`
	UserId   string `json:"userId"`
}

type UpdateContactRequest struct {
	Name     *string `json:"name,omitempty"`
	LastName *string `json:"lastName,omitempty"`
	Email    *string `json:"email,omitempty"`
	UserId   *string `json:"userId,omitempty"`
}
