package client

import (
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
