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

	GetUserByEmail(
		ctx context.Context,
		email string,
	) (*domain.UserDetail, error)
	RegisterPushToken(
		ctx context.Context,
		userID string,
		req domain.RegisterPushTokenRequest,
	) (*domain.AppDeviceResponse, error)
	DeletePushToken(
		ctx context.Context,
		userID string,
		appDeviceID int64,
	) error
}
