package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway-app/internal/domain"
	"net/http"
)

type HTTPDetectorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPDetectorClient(baseURL string) *HTTPDetectorClient {
	return &HTTPDetectorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: &loggingRoundTripper{next: http.DefaultTransport},
		},
	}
}

func (c *HTTPDetectorClient) GetDeviceEvents(ctx context.Context, deviceID string) (*[]domain.Event, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/devices/"+deviceID+"/events", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Detail string `json:"detail"`
		}
		detail := fmt.Sprintf("detector service returned status %d", resp.StatusCode)
		if json.NewDecoder(resp.Body).Decode(&errBody) == nil && errBody.Detail != "" {
			detail = errBody.Detail
		}
		return nil, &domain.DetectorError{
			Code:           domain.ErrCodeDetectorAnalysis,
			DetectorDetail: detail,
			StatusCode:     resp.StatusCode,
		}
	}

	return decodeResponse[[]domain.Event](resp)
}

func (c *HTTPDetectorClient) ClaimDevice(ctx context.Context, deviceID string) (*domain.DetectiveDevice, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/devices/"+deviceID+"/claim", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, detectiveError(resp)
	}

	return decodeResponse[domain.DetectiveDevice](resp)
}

func (c *HTTPDetectorClient) UnclaimDevice(ctx context.Context, deviceID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/devices/"+deviceID+"/unclaim", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return detectiveError(resp)
	}

	return nil
}

func (c *HTTPDetectorClient) GetDevice(ctx context.Context, deviceID string) (*domain.DetectiveDevice, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/devices/"+deviceID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, detectiveError(resp)
	}

	return decodeResponse[domain.DetectiveDevice](resp)
}

func (c *HTTPDetectorClient) UpdateDevice(ctx context.Context, deviceID string, req domain.UpdateDeviceRequest) (*domain.DetectiveDevice, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.baseURL+"/v1/devices/"+deviceID, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, detectiveError(resp)
	}

	return decodeResponse[domain.DetectiveDevice](resp)
}

func detectiveError(resp *http.Response) error {
	var errBody struct {
		Detail string `json:"detail"`
	}
	detail := fmt.Sprintf("detective-manager returned status %d", resp.StatusCode)
	if json.NewDecoder(resp.Body).Decode(&errBody) == nil && errBody.Detail != "" {
		detail = errBody.Detail
	}
	return &domain.DetectorError{
		Code:           domain.ErrCodeDetectorAnalysis,
		DetectorDetail: detail,
		StatusCode:     resp.StatusCode,
	}
}
