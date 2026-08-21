package client

import (
	"context"

	"gateway-app/internal/domain"
)

type DetectorClient interface {
	GetDeviceTracks(
		ctx context.Context,
		deviceID string,
	) (*[]domain.Track, error)

	GetTrackDetail(
		ctx context.Context,
		deviceID string,
		trackID string,
	) (*domain.TrackDetail, error)

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

	ListPersons(
		ctx context.Context,
		deviceID string,
	) (*[]domain.Person, error)

	CreatePerson(
		ctx context.Context,
		deviceID string,
		req domain.CreatePersonRequest,
	) (*domain.Person, error)

	GetPerson(
		ctx context.Context,
		deviceID string,
		personID string,
	) (*domain.Person, error)

	UpdatePerson(
		ctx context.Context,
		deviceID string,
		personID string,
		req domain.UpdatePersonRequest,
	) (*domain.Person, error)

	DeletePerson(
		ctx context.Context,
		deviceID string,
		personID string,
	) error

	AddPersonImages(
		ctx context.Context,
		deviceID string,
		personID string,
		req domain.AddPersonImagesRequest,
	) (*domain.Person, error)

	DeletePersonImage(
		ctx context.Context,
		deviceID string,
		personID string,
		imageID string,
	) error
}
