package domain

type PersonImage struct {
	ID       int    `json:"id"`
	ImageURL string `json:"imageUrl"`
}

type Person struct {
	ID       string        `json:"id"`
	DeviceID string        `json:"deviceId"`
	Name     string        `json:"name"`
	Role     string        `json:"role"`
	Images   []PersonImage `json:"images"`
}

type PersonWithContact struct {
	ID        string          `json:"id"`
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	Role      string          `json:"role"`
	Images    []PersonImage   `json:"images"`
	Contact   *ContactSummary `json:"contact"`
	IsOwner   bool            `json:"isOwner"`
}

type ContactSummary struct {
	Email     string `json:"email"`
	Cellphone string `json:"cellphone"`
	UserId    string `json:"userId"`
}

type ContactCreateInput struct {
	Email     string `json:"email" binding:"required"`
	Cellphone string `json:"cellphone"`
	UserId    string `json:"userId"`
}

type ContactUpdateInput struct {
	Email     *string `json:"email,omitempty"`
	Cellphone *string `json:"cellphone,omitempty"`
	UserId    *string `json:"userId,omitempty"`
}

type CreatePersonRequest struct {
	Name   string   `json:"name" binding:"required"`
	Role   string   `json:"role" binding:"required"`
	Images []string `json:"images"`
}

type AddPersonImagesRequest struct {
	Images []string `json:"images" binding:"required"`
}

type UpdatePersonRequest struct {
	Name *string `json:"name,omitempty"`
	Role *string `json:"role,omitempty"`
}

type CreatePersonWithContactRequest struct {
	FirstName string             `json:"firstName" binding:"required"`
	LastName  string             `json:"lastName" binding:"required"`
	Role      string             `json:"role" binding:"required"`
	Images    []string           `json:"images" binding:"required,min=3"`
	Contact   ContactCreateInput `json:"contact"`
}

type UpdatePersonWithContactRequest struct {
	FirstName *string             `json:"firstName,omitempty"`
	LastName  *string             `json:"lastName,omitempty"`
	Role      *string             `json:"role,omitempty"`
	Images    []string            `json:"images,omitempty"`
	Contact   *ContactUpdateInput `json:"contact,omitempty"`
}
