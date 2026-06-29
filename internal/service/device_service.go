package service

import (
	"context"
	"gateway-app/internal/client"
	"gateway-app/internal/domain"
)

type DeviceService struct {
	userClient      client.UserClient
	detectiveClient client.DetectorClient
}

func NewDeviceService(
	userClient client.UserClient,
	detectiveClient client.DetectorClient,
) DeviceService {
	return DeviceService{
		userClient:      userClient,
		detectiveClient: detectiveClient,
	}
}

func (s *DeviceService) GetMyDevice(
	ctx context.Context,
	userID string,
) (*domain.DeviceResponseDto, error) {

	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	detectiveDevice, err := s.detectiveClient.GetDevice(ctx, device.DeviceID)
	if err != nil {
		return nil, err
	}

	return domain.ToDeviceResponseDto(device, detectiveDevice), nil
}

func (s *DeviceService) LinkDevice(
	ctx context.Context,
	userID string,
	req domain.LinkDeviceRequest,
) (*domain.DeviceResponseDto, error) {

	device, err := s.userClient.LinkDevice(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	detectiveDevice, err := s.detectiveClient.ClaimDevice(ctx, device.DeviceID)
	if err != nil {
		return nil, err
	}

	return domain.ToDeviceResponseDto(device, detectiveDevice), nil
}

func (s *DeviceService) UnlinkDevice(
	ctx context.Context,
	userID string,
) error {

	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.userClient.UnlinkDevice(ctx, userID); err != nil {
		return err
	}

	return s.detectiveClient.UnclaimDevice(ctx, device.DeviceID)
}

func (s *DeviceService) UpdateMyDevice(
	ctx context.Context,
	userID string,
	req domain.UpdateDeviceRequest,
) (*domain.DeviceResponseDto, error) {

	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	detectiveDevice, err := s.detectiveClient.UpdateDevice(ctx, device.DeviceID, req)
	if err != nil {
		return nil, err
	}

	return domain.ToDeviceResponseDto(device, detectiveDevice), nil
}

func (s *DeviceService) GetMyDeviceEvents(
	ctx context.Context,
	userID string,
) (*[]domain.Event, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.GetDeviceEvents(ctx, device.DeviceID)
}

func (s *DeviceService) RegisterPushToken(
	ctx context.Context,
	userID string,
	req domain.RegisterPushTokenRequest,
) (*domain.AppDeviceResponse, error) {
	return s.userClient.RegisterPushToken(ctx, userID, req)
}
func (s *DeviceService) DeletePushToken(
	ctx context.Context,
	userID string,
	appDeviceID int64,
) error {
	return s.userClient.DeletePushToken(ctx, userID, appDeviceID)
}
