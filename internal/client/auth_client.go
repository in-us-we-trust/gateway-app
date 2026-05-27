package client

import (
	"context"
	"gateway-app/internal/domain"
)

type AuthClient interface {
	Register(
		ctx context.Context,
		req domain.AuthRegisterRequest,
	) error

	Login(
		ctx context.Context,
		req domain.AuthLoginRequest,
	) (*domain.LoginResponse, error)
}
