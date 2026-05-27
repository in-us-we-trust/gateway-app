package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway-app/internal/domain"
	"net/http"
)

type HTTPAuthClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPAuthClient(baseURL string) *HTTPAuthClient {
	return &HTTPAuthClient{
		baseURL: baseURL,
		client: &http.Client{
			Transport: &loggingRoundTripper{next: http.DefaultTransport},
		},
	}
}

func (c *HTTPAuthClient) Register(
	ctx context.Context,
	req domain.AuthRegisterRequest,
) error {

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/auth/register", c.baseURL)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return handleAPIError(resp)
	}

	return nil
}
