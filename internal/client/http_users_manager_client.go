package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway-app/internal/domain"
	"gateway-app/pkg/apierror"
	"net/http"
	"net/url"
)

type HTTPUserClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPUserClient(baseURL string) *HTTPUserClient {
	return &HTTPUserClient{
		baseURL: baseURL,
		client: &http.Client{
			Transport: &loggingRoundTripper{next: http.DefaultTransport},
		},
	}
}
func decodeResponse[T any](
	resp *http.Response,
) (*T, error) {

	var wrapper domain.StandardResponse[T]

	if err := json.NewDecoder(resp.Body).
		Decode(&wrapper); err != nil {

		return nil, err
	}

	return &wrapper.Data, nil
}

func decodeArrayResponse[T any](
	resp *http.Response,
) (*[]T, error) {
	var data []T

	if err := json.NewDecoder(resp.Body).
		Decode(&data); err != nil {

		return nil, err
	}

	return &data, nil
}

func (c *HTTPUserClient) GetUserDevice(
	ctx context.Context,
	userID string,
) (*domain.Device, error) {

	url := fmt.Sprintf(
		"%s/users/%s/device",
		c.baseURL,
		userID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	return decodeResponse[domain.Device](resp)
}

func (c *HTTPUserClient) LinkDevice(
	ctx context.Context,
	userID string,
	payload domain.LinkDeviceRequest,
) (*domain.Device, error) {

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"%s/users/%s/device",
		c.baseURL,
		userID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	return decodeResponse[domain.Device](resp)
}

func (c *HTTPUserClient) UnlinkDevice(
	ctx context.Context,
	userID string,
) error {

	url := fmt.Sprintf(
		"%s/users/%s/device",
		c.baseURL,
		userID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		url,
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return handleAPIError(resp)
	}

	return nil
}

func (c *HTTPUserClient) CreateUser(
	ctx context.Context,
	reqPayload domain.CreateUserRequest,
) (*domain.UserResponse, error) {
	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"%s/users",
		c.baseURL,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	user, err := decodeResponse[domain.User](resp)
	if err != nil {
		return nil, err
	}

	return &domain.UserResponse{Data: *user}, nil
}

func (c *HTTPUserClient) GetUserByEmail(
	ctx context.Context,
	email string,
) (*domain.UserDetail, error) {

	endpoint, err := url.Parse(
		fmt.Sprintf("%s/users/detail", c.baseURL),
	)
	if err != nil {
		return nil, err
	}

	q := endpoint.Query()
	q.Set("email", email)
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	return decodeResponse[domain.UserDetail](resp)
}

func (c *HTTPUserClient) RegisterPushToken(
	ctx context.Context,
	userID string,
	payload domain.RegisterPushTokenRequest,
) (*domain.AppDeviceResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/app-devices", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	var result domain.AppDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *HTTPUserClient) DeletePushToken(
	ctx context.Context,
	userID string,
	appDeviceID int64,
) error {
	url := fmt.Sprintf("%s/users/%s/app-devices/%d/push-token", c.baseURL, userID, appDeviceID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return handleAPIError(resp)
	}

	return nil
}

func (c *HTTPUserClient) ListContacts(
	ctx context.Context,
	deviceID string,
) (*[]domain.Contact, error) {
	url := fmt.Sprintf(
		"%s/devices/%s/contacts",
		c.baseURL,
		deviceID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	return decodeArrayResponse[domain.Contact](resp)
}

func (c *HTTPUserClient) CreateContact(
	ctx context.Context,
	deviceID string,
	payload domain.CreateContactRequest,
) (*domain.Contact, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"%s/devices/%s/contacts",
		c.baseURL,
		deviceID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, handleAPIError(resp)
	}

	return decodeObject[domain.Contact](resp)
}

func (c *HTTPUserClient) UpdateContact(
	ctx context.Context,
	deviceID string,
	contactID string,
	payload domain.UpdateContactRequest,
) (*domain.Contact, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"%s/devices/%s/contacts/%s",
		c.baseURL,
		deviceID,
		contactID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	return decodeObject[domain.Contact](resp)
}

func (c *HTTPUserClient) DeleteContact(
	ctx context.Context,
	deviceID string,
	contactID string,
) error {
	url := fmt.Sprintf(
		"%s/devices/%s/contacts/%s",
		c.baseURL,
		deviceID,
		contactID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		url,
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return handleAPIError(resp)
	}

	return nil
}

func (c *HTTPUserClient) GetContactByID(
	ctx context.Context,
	deviceID string,
	contactID string,
) (*domain.Contact, error) {
	url := fmt.Sprintf(
		"%s/devices/%s/contacts/%s",
		c.baseURL,
		deviceID,
		contactID,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	return decodeObject[domain.Contact](resp)
}

func handleAPIError(
	resp *http.Response,
) error {

	var apiErr apierror.APIError

	if err := json.NewDecoder(resp.Body).
		Decode(&apiErr); err != nil {

		return &domain.DetectorError{
			StatusCode:     resp.StatusCode,
			DetectorDetail: fmt.Sprintf("unexpected status: %d", resp.StatusCode),
		}
	}

	return &domain.DetectorError{
		StatusCode:     apiErr.Status,
		DetectorDetail: apiErr.Detail,
		Code:           apiErr.Code,
	}
}

func decodeObject[T any](resp *http.Response) (*T, error) {
	var data T

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
