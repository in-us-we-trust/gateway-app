package domain

type PersonImage struct {
	ID       int    `json:"id"`
	ImageURL string `json:"imageUrl"`
}

type Person struct {
	ID       int           `json:"id"`
	DeviceID string        `json:"deviceId"`
	Name     string        `json:"name"`
	Role     string        `json:"role"`
	Images   []PersonImage `json:"images"`
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

type FaceEmbedding struct {
	ID        int       `json:"id"`
	PersonID  int       `json:"personId"`
	Vector    []float64 `json:"vector"`
	CreatedAt string    `json:"createdAt"`
}
