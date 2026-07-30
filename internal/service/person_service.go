package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

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

func personWithContact(p *domain.Person, c *domain.Contact) *domain.PersonWithContact {
	firstName, lastName := splitName(p.Name)

	var cs *domain.ContactSummary
	var isOwner bool

	if c != nil {
		if c.Name != "" {
			firstName = c.Name
		}
		if c.LastName != "" {
			lastName = c.LastName
		}
		cs = &domain.ContactSummary{
			Email:     c.Email,
			Cellphone: c.Cellphone,
			UserId:    c.UserId,
		}
		isOwner = c.IsOwner
	}

	return &domain.PersonWithContact{
		ID:        p.ID,
		FirstName: firstName,
		LastName:  lastName,
		Role:      p.Role,
		Images:    p.Images,
		Contact:   cs,
		IsOwner:   isOwner,
	}
}

func splitName(name string) (firstName, lastName string) {
	parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
	firstName = parts[0]
	if len(parts) > 1 {
		lastName = strings.TrimSpace(parts[1])
	}
	return
}

func isNotFound(err error) bool {
	var detErr *domain.DetectorError
	return errors.As(err, &detErr) && detErr.StatusCode == http.StatusNotFound
}

func (s *PersonService) ListPersons(
	ctx context.Context,
	userID string,
) (*[]domain.PersonWithContact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	persons, err := s.detectiveClient.ListPersons(ctx, device.DeviceID)
	if err != nil {
		return nil, err
	}

	contacts, err := s.userClient.ListContacts(ctx, device.DeviceID)
	if err != nil && !isNotFound(err) {
		return nil, err
	}

	contactsByPersonID := make(map[string]*domain.Contact)
	if contacts != nil {
		for i := range *contacts {
			c := (*contacts)[i]
			if c.PersonId != "" {
				contactsByPersonID[c.PersonId] = &c
			}
		}
	}

	result := make([]domain.PersonWithContact, 0, len(*persons))
	for i := range *persons {
		p := (*persons)[i]
		result = append(result, *personWithContact(&p, contactsByPersonID[p.ID]))
	}

	return &result, nil
}

func (s *PersonService) CreatePerson(
	ctx context.Context,
	userID string,
	req domain.CreatePersonWithContactRequest,
) (*domain.PersonWithContact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.FirstName + " " + req.LastName)
	person, err := s.detectiveClient.CreatePerson(ctx, device.DeviceID, domain.CreatePersonRequest{
		Name:   name,
		Role:   req.Role,
		Images: req.Images,
	})
	if err != nil {
		return nil, err
	}

	contactReq := domain.CreateContactRequest{
		Name:      req.FirstName,
		LastName:  req.LastName,
		Email:     req.Contact.Email,
		Cellphone: req.Contact.Cellphone,
		UserId:    req.Contact.UserId,
		PersonId:  person.ID,
	}
	contact, err := s.userClient.CreateContact(ctx, device.DeviceID, contactReq)
	if err != nil {
		return nil, err
	}

	return personWithContact(person, contact), nil
}

func (s *PersonService) GetPerson(
	ctx context.Context,
	userID string,
	personID string,
) (*domain.PersonWithContact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	person, err := s.detectiveClient.GetPerson(ctx, device.DeviceID, personID)
	if err != nil {
		return nil, err
	}

	contact, err := s.userClient.GetContactByPersonID(ctx, device.DeviceID, personID)
	if err != nil && !isNotFound(err) {
		return nil, err
	}

	return personWithContact(person, contact), nil
}

func (s *PersonService) UpdatePerson(
	ctx context.Context,
	userID string,
	personID string,
	req domain.UpdatePersonWithContactRequest,
) (*domain.PersonWithContact, error) {
	device, err := s.userClient.GetUserDevice(ctx, userID)
	if err != nil {
		return nil, err
	}

	personReq := domain.UpdatePersonRequest{Role: req.Role}
	if req.FirstName != nil || req.LastName != nil {
		name := buildPersonName(req.FirstName, req.LastName)
		personReq.Name = &name
	}

	person, err := s.detectiveClient.UpdatePerson(ctx, device.DeviceID, personID, personReq)
	if err != nil {
		return nil, err
	}

	var contact *domain.Contact

	if req.Contact != nil {
		contactReq := domain.UpdateContactRequest{
			Email:     req.Contact.Email,
			Cellphone: req.Contact.Cellphone,
			UserId:    req.Contact.UserId,
			Name:      req.FirstName,
			LastName:  req.LastName,
		}
		contact, err = s.userClient.UpdateContactByPersonID(ctx, device.DeviceID, personID, contactReq)
		if err != nil {
			return nil, err
		}
	} else {
		contact, err = s.userClient.GetContactByPersonID(ctx, device.DeviceID, personID)
		if err != nil && !isNotFound(err) {
			return nil, err
		}
	}

	return personWithContact(person, contact), nil
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

	personErr := s.detectiveClient.DeletePerson(ctx, device.DeviceID, personID)
	contactErr := s.userClient.DeleteContactByPersonID(ctx, device.DeviceID, personID)

	if personErr != nil && !isNotFound(personErr) {
		return personErr
	}
	if contactErr != nil && !isNotFound(contactErr) {
		return contactErr
	}

	return nil
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

func buildPersonName(firstName, lastName *string) string {
	var f, l string
	if firstName != nil {
		f = *firstName
	}
	if lastName != nil {
		l = *lastName
	}
	return strings.TrimSpace(f + " " + l)
}
