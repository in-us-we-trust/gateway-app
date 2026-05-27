package client

import (
	"context"
	"gateway-app/internal/domain"
)

type UserClient interface {
	GetUserDevice(
		ctx context.Context,
		userID string,
	) (*domain.Device, error)

	LinkDevice(
		ctx context.Context,
		userID string,
		req domain.LinkDeviceRequest,
	) (*domain.Device, error)

	UnlinkDevice(
		ctx context.Context,
		userID string,
	) error
	CreateUser(
		ctx context.Context,
		req domain.CreateUserRequest,
	) (*domain.UserResponse, error)
}
