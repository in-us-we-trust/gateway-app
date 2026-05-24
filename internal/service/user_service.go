package service

import (
	"context"
	"gateway-app/internal/client"
	"gateway-app/internal/domain"
)

type UserService struct {
	userClient      client.UserClient
	detectiveClient client.DetectorClient
}

func NewUserService(
	userClient client.UserClient,
	detectiveClient client.DetectorClient,
) UserService {
	return UserService{
		userClient:      userClient,
		detectiveClient: detectiveClient,
	}
}

func (s *UserService) GetMyDevice(
	ctx context.Context,
	userID string,
) (*domain.Device, error) {

	return s.userClient.GetUserDevice(
		ctx,
		userID,
	)
}

func (s *UserService) LinkDevice(
	ctx context.Context,
	userID string,
	req domain.LinkDeviceRequest,
) (*domain.Device, error) {

	return s.userClient.LinkDevice(
		ctx,
		userID,
		req,
	)
}

func (s *UserService) UnlinkDevice(
	ctx context.Context,
	userID string,
) error {

	return s.userClient.UnlinkDevice(
		ctx,
		userID,
	)
}

func (s *UserService) GetMyDeviceEvents(
	ctx context.Context,
	userID string,
) (*[]domain.Event, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.GetDeviceEvents(ctx, device.DeviceID)
}
