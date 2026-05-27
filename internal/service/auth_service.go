package service

import (
	"context"
	"gateway-app/internal/client"
	"gateway-app/internal/domain"
)

type AuthService struct {
	userClient client.UserClient
	authClient client.AuthClient
}

func NewAuthService(
	userClient client.UserClient,
	authClient client.AuthClient,
) AuthService {
	return AuthService{
		userClient: userClient,
		authClient: authClient,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req domain.RegisterRequest,
) (*domain.UserResponse, error) {
	if err := s.authClient.Register(ctx, domain.AuthRegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		return nil, err
	}

	return s.userClient.CreateUser(ctx, domain.CreateUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
}

func (s *AuthService) Login(
	ctx context.Context,
	req domain.LoginRequest,
) (*domain.LoginResponse, error) {
	return s.authClient.Login(ctx, domain.AuthLoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
}