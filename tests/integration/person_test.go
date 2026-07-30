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
	testPersonID = "person-uuid-1"
	testImageID  = "5"
)

var testPerson = domain.Person{
	ID:       testPersonID,
	DeviceID: testDeviceID,
	Name:     "Juan Perez",
	Role:     "Cuidador",
	Images: []domain.PersonImage{
		{ID: 5, ImageURL: "http://img/face.jpg"},
	},
}

var testPersonContact = domain.Contact{
	ContactId: "contact-uuid-1",
	PersonId:  testPersonID,
	UserId:    "user-uuid-1",
	Name:      "Juan",
	LastName:  "Perez",
	Email:     "juan@example.com",
	Cellphone: "+5491123456789",
	IsOwner:   false,
}

func setupPersonRouter(usersURL, detectiveURL string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	userClient := client.NewHTTPUserClient(usersURL)
	detectiveClient := client.NewHTTPDetectorClient(detectiveURL)
	personService := service.NewPersonService(userClient, detectiveClient)
	personHandler := handlers.NewPersonHandler(personService)

	router := gin.New()
	protected := router.Group("/api/v1")
	protected.Use(middleware.JWTAuth())
	{
		protected.GET("/me/device/persons", personHandler.ListPersons)
		protected.POST("/me/device/persons", personHandler.CreatePerson)
		protected.GET("/me/device/persons/:personId", personHandler.GetPerson)
		protected.PATCH("/me/device/persons/:personId", personHandler.UpdatePerson)
		protected.DELETE("/me/device/persons/:personId", personHandler.DeletePerson)
		protected.POST("/me/device/persons/:personId/images", personHandler.AddPersonImages)
		protected.DELETE("/me/device/persons/:personId/images/:imageId", personHandler.DeletePersonImage)
	}

	return router
}

func usersMuxWithDevice(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, usersManagerDevice))
	})
	return mux
}

func contactEndpoint(personID string) string {
	return fmt.Sprintf("/devices/%s/persons/%s/contact", testDeviceID, personID)
}

func contactsListEndpoint() string {
	return fmt.Sprintf("/devices/%s/contacts", testDeviceID)
}

func TestListPersons_ReturnsPersonsWithContact(t *testing.T) {
	persons := []domain.Person{testPerson}
	contacts := []domain.Contact{testPersonContact}

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, persons))
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactsListEndpoint(), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		b, _ := json.Marshal(contacts)
		w.Write(b)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].([]interface{})

	require.Len(t, data, 1)
	first := data[0].(map[string]interface{})
	assert.Equal(t, "Juan", first["firstName"])
	assert.Equal(t, "Perez", first["lastName"])
	assert.Equal(t, "Cuidador", first["role"])

	contact := first["contact"].(map[string]interface{})
	assert.Equal(t, "juan@example.com", contact["email"])
	assert.Equal(t, "+5491123456789", contact["cellphone"])
}

func TestListPersons_ContactNotFound_ReturnsPersonWithNullContact(t *testing.T) {
	persons := []domain.Person{testPerson}

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, persons))
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactsListEndpoint(), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "no contacts"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].([]interface{})

	require.Len(t, data, 1)
	first := data[0].(map[string]interface{})

	assert.Equal(t, "Juan", first["firstName"])
	assert.Equal(t, "Perez", first["lastName"])
	assert.Nil(t, first["contact"])
}

func TestListPersons_UsersManagerError_ReturnsError(t *testing.T) {
	usersMux := http.NewServeMux()
	usersMux.HandleFunc(fmt.Sprintf("/users/%s/device", testUserID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "no device linked"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, http.NewServeMux())

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListPersons_DetectiveManagerError_ReturnsError(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "internal error"})
	})

	usersMux := usersMuxWithDevice(t)
	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetPerson_ReturnsPersonWithContact(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, testPerson))
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testPersonContact)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].(map[string]interface{})

	assert.Equal(t, testPersonID, data["id"])
	assert.Equal(t, "Juan", data["firstName"])
	assert.Equal(t, "Perez", data["lastName"])

	contact := data["contact"].(map[string]interface{})
	assert.Equal(t, "juan@example.com", contact["email"])
	assert.Equal(t, "+5491123456789", contact["cellphone"])
	assert.Equal(t, "user-uuid-1", contact["userId"])
}

func TestGetPerson_ContactNotFound_ReturnsPersonWithNullContact(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, testPerson))
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "contact not found"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].(map[string]interface{})

	assert.Equal(t, "Juan", data["firstName"])
	assert.Equal(t, "Perez", data["lastName"])
	assert.Nil(t, data["contact"])
}

func TestGetPerson_PersonNotFound_ReturnsNotFound(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "person not found"})
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreatePerson_PropagatesPersonIdToContact(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		var req domain.CreatePersonRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "Juan Perez", req.Name)
		assert.Equal(t, "Cuidador", req.Role)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, testPerson))
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactsListEndpoint(), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		var req domain.CreateContactRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		assert.Equal(t, testPersonID, req.PersonId)
		assert.Equal(t, "Juan", req.Name)
		assert.Equal(t, "Perez", req.LastName)
		assert.Equal(t, "juan@example.com", req.Email)
		assert.Equal(t, "+5491123456789", req.Cellphone)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(testPersonContact)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.CreatePersonWithContactRequest{
		FirstName: "Juan",
		LastName:  "Perez",
		Role:      "Cuidador",
		Images:    []string{"base64img1", "base64img2", "base64img3"},
		Contact: domain.ContactCreateInput{
			Email:     "juan@example.com",
			Cellphone: "+5491123456789",
			UserId:    "user-uuid-1",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].(map[string]interface{})

	assert.Equal(t, "Juan", data["firstName"])
	assert.Equal(t, "Perez", data["lastName"])
	assert.Equal(t, "Cuidador", data["role"])
	assert.NotNil(t, data["contact"])
}

func TestCreatePerson_InvalidBody_ReturnsBadRequest(t *testing.T) {
	router := setupPersonRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePerson_TooFewImages_ReturnsBadRequest(t *testing.T) {
	router := setupPersonRouter("http://unused", "http://unused")

	reqBody, _ := json.Marshal(domain.CreatePersonWithContactRequest{
		FirstName: "Juan",
		LastName:  "Perez",
		Role:      "Cuidador",
		Images:    []string{"only-one"},
		Contact:   domain.ContactCreateInput{Email: "a@b.com"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePerson_DetectiveManagerError_ReturnsError(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "invalid images"})
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.CreatePersonWithContactRequest{
		FirstName: "Juan",
		LastName:  "Perez",
		Role:      "Cuidador",
		Images:    []string{"img1", "img2", "img3"},
		Contact:   domain.ContactCreateInput{Email: "a@b.com"},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePerson_UpdatesPersonAndContact(t *testing.T) {
	newFirstName := "Juan Updated"
	newEmail := "updated@example.com"

	updatedPerson := testPerson
	updatedPerson.Name = newFirstName

	updatedContact := testPersonContact
	updatedContact.Name = newFirstName
	updatedContact.Email = newEmail

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)

		var req domain.UpdatePersonRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, newFirstName, *req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, updatedPerson))
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)

		var req domain.UpdateContactRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, newEmail, *req.Email)

		assert.Equal(t, newFirstName, *req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedContact)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.UpdatePersonWithContactRequest{
		FirstName: &newFirstName,
		Contact: &domain.ContactUpdateInput{
			Email: &newEmail,
		},
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device/persons/"+testPersonID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].(map[string]interface{})

	assert.Equal(t, "Juan Updated", data["firstName"])

	contact := data["contact"].(map[string]interface{})
	assert.Equal(t, newEmail, contact["email"])
}

func TestUpdatePerson_InvalidBody_ReturnsBadRequest(t *testing.T) {
	router := setupPersonRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device/persons/"+testPersonID, bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePerson_DeletesBothPersonAndContact(t *testing.T) {
	personDeleted := false
	contactDeleted := false

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		personDeleted = true
		w.WriteHeader(http.StatusNoContent)
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		contactDeleted = true
		w.WriteHeader(http.StatusNoContent)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, personDeleted)
	assert.True(t, contactDeleted)
}

func TestDeletePerson_ContactNotFound_ReturnsNoContent(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": 404, "detail": "contact not found"})
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeletePerson_PersonNotFound_ReturnsNoContent(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "person not found"})
	})

	contactDeleted := false
	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		contactDeleted = true
		w.WriteHeader(http.StatusNoContent)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, contactDeleted)
}

func TestDeletePerson_PersonNonNotFoundError_ReturnsError(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "internal error"})
	})

	usersMux := usersMuxWithDevice(t)
	usersMux.HandleFunc(contactEndpoint(testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	usersSrv := newUsersMock(t, usersMux)
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAddPersonImages_ReturnsUpdatedPerson(t *testing.T) {
	updatedPerson := testPerson
	updatedPerson.Images = append(updatedPerson.Images, domain.PersonImage{ID: 6, ImageURL: "http://img/face2.jpg"})

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s/images", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		var req domain.AddPersonImagesRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		require.Len(t, req.Images, 1)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, updatedPerson))
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.AddPersonImagesRequest{Images: []string{"base64newface"}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons/"+testPersonID+"/images", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].(map[string]interface{})
	images := data["images"].([]interface{})

	assert.Len(t, images, 2)
}

func TestAddPersonImages_InvalidBody_ReturnsBadRequest(t *testing.T) {
	router := setupPersonRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons/"+testPersonID+"/images", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePersonImage_ReturnsNoContent(t *testing.T) {
	deleteCalled := false

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s/images/%s", testDeviceID, testPersonID, testImageID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		deleteCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID+"/images/"+testImageID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, deleteCalled)
}

func TestDeletePersonImage_DetectiveManagerError_ReturnsError(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s/images/%s", testDeviceID, testPersonID, testImageID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "image not found"})
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID+"/images/"+testImageID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
