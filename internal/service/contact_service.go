package service

import (
	"context"

	"gateway-app/internal/client"
	"gateway-app/internal/domain"
)

type ContactService struct {
	userClient      client.UserClient
	detectiveClient client.DetectorClient
}

func NewContactService(
	userClient client.UserClient,
	detectiveClient client.DetectorClient,
) ContactService {
	return ContactService{
		userClient:      userClient,
		detectiveClient: detectiveClient,
	}
}

func (s *ContactService) ListContacts(
	ctx context.Context,
	userID string,
) (*[]domain.Contact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.userClient.ListContacts(ctx, device.DeviceID)
}

func (s *ContactService) CreateContact(
	ctx context.Context,
	userID string,
	req domain.CreateContactRequest,
) (*domain.Contact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.userClient.CreateContact(ctx, device.DeviceID, req)
}

func (s *ContactService) UpdateContact(
	ctx context.Context,
	userID string,
	contactID string,
	req domain.UpdateContactRequest,
) (*domain.Contact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.userClient.UpdateContact(ctx, device.DeviceID, contactID, req)
}

func (s *ContactService) DeleteContact(
	ctx context.Context,
	userID string,
	contactID string,
) error {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return err
	}

	return s.userClient.DeleteContact(ctx, device.DeviceID, contactID)
}

func (s *ContactService) GetContactByID(
	ctx context.Context,
	userID string,
	contactID string,
) (*domain.Contact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	contacts, err := s.userClient.GetContactByID(ctx, device.DeviceID, contactID)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}
