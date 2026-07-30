package domain

type NotificationToken struct {
	AppDeviceId int    `json:"appDeviceId"`
	DeviceName  string `json:"deviceName"`
	Platform    string `json:"platform"`
	Token       string `json:"token"`
}

type Contact struct {
	ContactId          string              `json:"contactId"`
	PersonId           string              `json:"personId"`
	UserId             string              `json:"userId"`
	Name               string              `json:"name"`
	LastName           string              `json:"lastName"`
	Email              string              `json:"email"`
	Cellphone          string              `json:"cellphone"`
	IsOwner            bool                `json:"isOwner"`
	NotificationTokens []NotificationToken `json:"notificationTokens"`
}

type CreateContactRequest struct {
	Name      string `json:"name"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	UserId    string `json:"userId"`
	PersonId  string `json:"personId"`
	Cellphone string `json:"cellphone"`
}

type UpdateContactRequest struct {
	Name      *string `json:"name,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
	Email     *string `json:"email,omitempty"`
	UserId    *string `json:"userId,omitempty"`
	Cellphone *string `json:"cellphone,omitempty"`
}
