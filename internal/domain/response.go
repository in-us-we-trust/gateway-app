package domain

type StandardResponse[T any] struct {
	Data T `json:"data"`
}
