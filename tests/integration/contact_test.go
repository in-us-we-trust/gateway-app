package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	testContactID = "contact-uuid-1"
)

var testContact = domain.Contact{
	ContactId: testContactID,
	UserId:    "user-uuid-1",
	Name:      "John",
	LastName:  "Doe",
	Email:     "john@example.com",
	IsOwner:   false,
	NotificationTokens: []domain.NotificationToken{
		{
			AppDeviceId: 1,
			DeviceName:  "iPhone",
			Platform:    "ios",
			Token:       "fcm-token-123",
		},
	},
}

func setupContactRouter(usersURL, detectiveURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userClient := client.NewHTTPUserClient(usersURL)
	detectiveClient := client.NewHTTPDetectorClient(detectiveURL)
	contactService := service.NewContactService(userClient, detectiveClient)
	contactHandler := handlers.NewContactHandler(contactService)

	router := gin.New()
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/me/device/contacts", contactHandler.ListContacts)
		protected.POST("/me/device/contacts", contactHandler.CreateContact)
		protected.PATCH("/me/device/contacts/:contactId", contactHandler.UpdateContact)
		protected.DELETE("/me/device/contacts/:contactId", contactHandler.DeleteContact)
	}

	return router
}

func TestListContacts_ReturnsContacts(t *testing.T) {
	contacts := []domain.Contact{testContact}

	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	usersMux.HandleFunc(fmt.Sprintf("/devices/%s/contacts", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(contacts)
		w.Write(b)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/contacts", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].([]interface{})

	require.Len(t, data, 1)
	first := data[0].(map[string]interface{})
	assert.Equal(t, "John", first["name"])
	assert.Equal(t, "Doe", first["lastName"])
	assert.Equal(t, "john@example.com", first["email"])
}

func TestListContacts_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "no device linked"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/contacts", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateContact_ReturnsContact(t *testing.T) {
	usersMux := http.NewServeMux()

	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	usersMux.HandleFunc(fmt.Sprintf("/devices/%s/contacts", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		var req domain.CreateContactRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		assert.Equal(t, "John", req.Name)
		assert.Equal(t, "Doe", req.LastName)
		assert.Equal(t, "john@example.com", req.Email)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		require.NoError(t, json.NewEncoder(w).Encode(testContact))
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.CreateContactRequest{
		Name:     "John",
		LastName: "Doe",
		Email:    "john@example.com",
	})

	req := httptest.NewRequest(http.MethodPost,
		"/api/v1/me/device/contacts",
		bytes.NewBuffer(reqBody),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))

	data := respBody["data"].(map[string]interface{})

	assert.Equal(t, "John", data["name"])
	assert.Equal(t, "Doe", data["lastName"])
	assert.Equal(t, "john@example.com", data["email"])
}

func TestCreateContact_InvalidBody_ReturnsBadRequest(t *testing.T) {
	router := setupContactRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/contacts", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateContact_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	usersMux.HandleFunc(fmt.Sprintf("/devices/%s/contacts", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 400, "detail": "invalid contact"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.CreateContactRequest{
		Name:     "John",
		LastName: "Doe",
		Email:    "john@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/contacts", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateContact_ReturnsUpdatedContact(t *testing.T) {
	newName := "Jane"
	updated := testContact
	updated.Name = newName

	usersMux := http.NewServeMux()

	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	usersMux.HandleFunc(fmt.Sprintf("/devices/%s/contacts/%s", testDeviceID, testContactID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)

		var req domain.UpdateContactRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		assert.Equal(t, &newName, req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		require.NoError(t, json.NewEncoder(w).Encode(updated))
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.UpdateContactRequest{
		Name: &newName,
	})

	req := httptest.NewRequest(http.MethodPatch,
		"/api/v1/me/device/contacts/"+testContactID,
		bytes.NewBuffer(reqBody),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))

	data := respBody["data"].(map[string]interface{})

	assert.Equal(t, newName, data["name"])
}

func TestUpdateContact_InvalidBody_ReturnsBadRequest(t *testing.T) {
	router := setupContactRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device/contacts/"+testContactID, bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteContact_ReturnsNoContent(t *testing.T) {
	deleteCalled := false

	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	usersMux.HandleFunc(fmt.Sprintf("/devices/%s/contacts/%s", testDeviceID, testContactID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		deleteCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/contacts/"+testContactID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, deleteCalled)
}

func TestDeleteContact_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})

	usersMux.HandleFunc(fmt.Sprintf("/devices/%s/contacts/%s", testDeviceID, testContactID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "contact not found"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupContactRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/contacts/"+testContactID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
