package service

import (
	"context"

	"gateway-app/internal/client"
	"gateway-app/internal/domain"
)

type PersonService struct {
	userClient      client.UserClient
	detectiveClient client.DetectorClient
}

func NewPersonService(
	userClient client.UserClient,
	detectiveClient client.DetectorClient,
) PersonService {
	return PersonService{
		userClient:      userClient,
		detectiveClient: detectiveClient,
	}
}

func (s *PersonService) ListPersons(
	ctx context.Context,
	userID string,
) (*[]domain.Person, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.ListPersons(ctx, device.DeviceID)
}

func (s *PersonService) CreatePerson(
	ctx context.Context,
	userID string,
	req domain.CreatePersonRequest,
) (*domain.Person, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.CreatePerson(ctx, device.DeviceID, req)
}

func (s *PersonService) GetPerson(
	ctx context.Context,
	userID string,
	personID string,
) (*domain.Person, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.GetPerson(ctx, device.DeviceID, personID)
}

func (s *PersonService) UpdatePerson(
	ctx context.Context,
	userID string,
	personID string,
	req domain.UpdatePersonRequest,
) (*domain.Person, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.UpdatePerson(ctx, device.DeviceID, personID, req)
}

func (s *PersonService) DeletePerson(
	ctx context.Context,
	userID string,
	personID string,
) error {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return err
	}

	return s.detectiveClient.DeletePerson(ctx, device.DeviceID, personID)
}

func (s *PersonService) AddPersonImages(
	ctx context.Context,
	userID string,
	personID string,
	req domain.AddPersonImagesRequest,
) (*domain.Person, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.AddPersonImages(ctx, device.DeviceID, personID, req)
}

func (s *PersonService) GetFaceEmbeddings(
	ctx context.Context,
	userID string,
	personID string,
) (*[]domain.FaceEmbedding, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.detectiveClient.GetFaceEmbeddings(ctx, device.DeviceID, personID)
}

func (s *PersonService) DeletePersonImage(
	ctx context.Context,
	userID string,
	personID string,
	imageID string,
) error {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return err
	}

	return s.detectiveClient.DeletePersonImage(ctx, device.DeviceID, personID, imageID)
}
