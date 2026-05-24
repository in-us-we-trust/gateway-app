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
}
