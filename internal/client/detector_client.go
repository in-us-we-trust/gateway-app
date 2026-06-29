package client

import (
	"context"

	"gateway-app/internal/domain"
)

type DetectorClient interface {
	GetDeviceEvents(
		ctx context.Context,
		deviceID string,
	) (*[]domain.Event, error)

	ClaimDevice(
		ctx context.Context,
		deviceID string,
	) (*domain.DetectiveDevice, error)

	UnclaimDevice(
		ctx context.Context,
		deviceID string,
	) error

	GetDevice(
		ctx context.Context,
		deviceID string,
	) (*domain.DetectiveDevice, error)

	UpdateDevice(
		ctx context.Context,
		deviceID string,
		req domain.UpdateDeviceRequest,
	) (*domain.DetectiveDevice, error)
}
