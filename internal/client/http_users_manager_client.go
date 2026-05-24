package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway-app/internal/domain"
	"gateway-app/pkg/apierror"
	"net/http"
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

func handleAPIError(
	resp *http.Response,
) error {

	var apiErr apierror.APIError

	if err := json.NewDecoder(resp.Body).
		Decode(&apiErr); err != nil {

		return fmt.Errorf(
			"unexpected status: %d",
			resp.StatusCode,
		)
	}

	return fmt.Errorf(
		"external api error (%d): %s",
		apiErr.Status,
		apiErr.Detail,
	)
}
