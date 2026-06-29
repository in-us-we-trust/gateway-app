package integration_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway-app/internal/client"
	"gateway-app/internal/domain"
	"gateway-app/internal/handlers"
	"gateway-app/internal/middleware"
	"gateway-app/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUserID   = "user-123"
	testDeviceID = "device-456"
)

func buildBearerToken(t *testing.T, userID string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]interface{}{
		"user": map[string]string{"id": userID},
	})
	require.NoError(t, err)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return "Bearer eyJhbGciOiJub25lIn0." + encoded + ".sig"
}

var usersManagerDevice = domain.Device{
	DeviceID:  testDeviceID,
	Name:      "Test Device",
	CreatedAt: "2026-06-27T17:00:00Z",
}

var detectiveDevice = domain.DetectiveDevice{
	ID:                 testDeviceID,
	SerialNumber:       "ESP32-000001",
	City:               "San Miguel",
	LastSeen:           "2026-06-27T18:30:00Z",
	LastBoot:           "2026-06-27T17:45:00Z",
	Status:             "CLAIMED",
	Uptime:             2700,
	DetectionThreshold: 85,
	Location:           "Main Entrance",
	FirmwareVersion:    "v1.3.7",
}

func newUsersMock(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func newDetectiveMock(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func encodeStdResponse(t *testing.T, data interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(domain.StandardResponse[interface{}]{Data: data})
	require.NoError(t, err)
	return b
}

func setupRouter(usersURL, detectiveURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userClient := client.NewHTTPUserClient(usersURL)
	detectiveClient := client.NewHTTPDetectorClient(detectiveURL)
	deviceService := service.NewDeviceService(userClient, detectiveClient)
	deviceHandler := handlers.NewDeviceHandler(deviceService)

	router := gin.New()
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/me/device", deviceHandler.GetMyDevice)
		protected.POST("/me/device", deviceHandler.LinkMyDevice)
		protected.PATCH("/me/device", deviceHandler.UpdateMyDevice)
		protected.DELETE("/me/device", deviceHandler.UnlinkMyDevice)
		protected.GET("/me/device/events", deviceHandler.GetMyDeviceEvents)
	}

	return router
}

func TestLinkDevice_EnrichesResponseFromDetectiveManager(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/claim", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, detectiveDevice))
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.LinkDeviceRequest{DeviceID: testDeviceID, Name: "Test Device"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	resp := respBody["data"].(map[string]interface{})

	assert.Equal(t, testDeviceID, resp["deviceId"])
	assert.Equal(t, "Test Device", resp["name"])
	assert.Equal(t, "ESP32-000001", resp["serialNumber"])
	assert.Equal(t, "San Miguel", resp["city"])
	assert.Equal(t, "2026-06-27T18:30:00Z", resp["lastSeen"])
	assert.Equal(t, "2026-06-27T17:45:00Z", resp["lastBoot"])
	assert.Equal(t, "CLAIMED", resp["status"])
	assert.Equal(t, float64(85), resp["detectionThreshold"])
	assert.Equal(t, "Main Entrance", resp["location"])
	assert.Equal(t, "v1.3.7", resp["firmwareVersion"])
}

func TestLinkDevice_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404,
			"detail": "device not found",
			"code":   "NOT_FOUND",
		})
	})

	detectiveMux := http.NewServeMux()
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	body, _ := json.Marshal(domain.LinkDeviceRequest{DeviceID: testDeviceID, Name: "Test Device"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUnlinkDevice_CallsUnclaimOnDetectiveManager(t *testing.T) {
	unclaimCalled := false

	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(encodeStdResponse(t, usersManagerDevice))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/unclaim", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		unclaimCalled = true
		w.WriteHeader(http.StatusOK)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, unclaimCalled, "expected detective-manager unclaim endpoint to be called")
}

func TestUnlinkDevice_UsersManagerDeleteError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(encodeStdResponse(t, usersManagerDevice))
		case http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": 404,
				"detail": "device not found",
			})
		}
	})

	detectiveMux := http.NewServeMux()
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDevice_EnrichesResponseFromDetectiveManager(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/"+testDeviceID) {
			require.Equal(t, http.MethodGet, r.Method)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(encodeStdResponse(t, detectiveDevice))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	resp := body["data"].(map[string]interface{})

	assert.Equal(t, testDeviceID, resp["deviceId"])
	assert.Equal(t, "Test Device", resp["name"])
	assert.Equal(t, "2026-06-27T17:00:00Z", resp["createdAt"])
	assert.Equal(t, "ESP32-000001", resp["serialNumber"])
	assert.Equal(t, "San Miguel", resp["city"])
	assert.Equal(t, "2026-06-27T18:30:00Z", resp["lastSeen"])
	assert.Equal(t, "2026-06-27T17:45:00Z", resp["lastBoot"])
	assert.Equal(t, "CLAIMED", resp["status"])
	assert.Equal(t, float64(2700), resp["uptime"])
	assert.Equal(t, float64(85), resp["detectionThreshold"])
	assert.Equal(t, "Main Entrance", resp["location"])
	assert.Equal(t, "v1.3.7", resp["firmwareVersion"])
}

func TestGetDevice_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404,
			"detail": "no device linked",
		})
	})

	detectiveMux := http.NewServeMux()
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDevice_DetectiveManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"detail": "internal error",
		})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateDevice_UpdatesAndEnrichesResponse(t *testing.T) {
	city := "Buenos Aires"
	threshold := 90
	location := "Back Door"

	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	updatedDetectiveDevice := detectiveDevice
	updatedDetectiveDevice.City = city
	updatedDetectiveDevice.DetectionThreshold = threshold
	updatedDetectiveDevice.Location = location

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/"+testDeviceID) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		require.Equal(t, http.MethodPatch, r.Method)

		var req domain.UpdateDeviceRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, &city, req.City)
		assert.Equal(t, &threshold, req.DetectionThreshold)
		assert.Equal(t, &location, req.Location)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, updatedDetectiveDevice))
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.UpdateDeviceRequest{City: &city, DetectionThreshold: &threshold, Location: &location})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	resp := respBody["data"].(map[string]interface{})

	assert.Equal(t, testDeviceID, resp["deviceId"])
	assert.Equal(t, "Test Device", resp["name"])
	assert.Equal(t, city, resp["city"])
	assert.Equal(t, float64(threshold), resp["detectionThreshold"])
	assert.Equal(t, location, resp["location"])
}

func TestUpdateDevice_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": 404,
			"detail": "no device linked",
		})
	})

	detectiveMux := http.NewServeMux()
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	city := "Buenos Aires"
	reqBody, _ := json.Marshal(domain.UpdateDeviceRequest{City: &city})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateDevice_DetectiveManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"detail": "invalid field",
		})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	location := "Garage"
	reqBody, _ := json.Marshal(domain.UpdateDeviceRequest{Location: &location})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnyEndpoint_MissingAuthorizationHeader_ReturnsUnauthorized(t *testing.T) {
	router := setupRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnyEndpoint_MalformedBearerToken_ReturnsUnauthorized(t *testing.T) {
	router := setupRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt.token.with.too.many.parts")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnyEndpoint_TokenMissingUserID_ReturnsUnauthorized(t *testing.T) {
	payload, _ := json.Marshal(map[string]interface{}{"sub": "someone"})
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	token := "Bearer eyJhbGciOiJub25lIn0." + encoded + ".sig"

	router := setupRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLinkDevice_InvalidRequestBody_ReturnsBadRequest(t *testing.T) {
	router := setupRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateDevice_InvalidRequestBody_ReturnsBadRequest(t *testing.T) {
	router := setupRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLinkDevice_DetectiveManagerClaimError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/claim", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "device already claimed"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.LinkDeviceRequest{DeviceID: testDeviceID, Name: "Test Device"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestUnlinkDevice_GetUserDeviceError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "no device linked"})
	})

	detectiveMux := http.NewServeMux()
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUnlinkDevice_DetectiveManagerUnclaimError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(encodeStdResponse(t, usersManagerDevice))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/unclaim", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "unclaim failed"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetDeviceEvents_ReturnsEvents(t *testing.T) {
	events := []domain.Event{
		{ID: "evt-1", DeviceID: testDeviceID, Timestamp: "2026-06-27T18:00:00Z", Intruder: true, ImageURL: "http://img/1.jpg", EventType: "motion"},
		{ID: "evt-2", DeviceID: testDeviceID, Timestamp: "2026-06-27T19:00:00Z", Intruder: false, ImageURL: "http://img/2.jpg", EventType: "bell"},
	}

	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/events", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, events))
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/events", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].([]interface{})

	require.Len(t, data, 2)
	first := data[0].(map[string]interface{})
	assert.Equal(t, "evt-1", first["id"])
	assert.Equal(t, "motion", first["eventType"])
}

func TestGetDeviceEvents_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "no device linked"})
	})

	detectiveMux := http.NewServeMux()
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/events", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDeviceEvents_DetectiveManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/events", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "events fetch failed"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/events", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
