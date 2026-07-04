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
	testPersonID = "1"
	testImageID  = "5"
)

var testPerson = domain.Person{
	ID:       1,
	DeviceID: testDeviceID,
	Name:     "Juan Perez",
	Role:     "Cuidador",
	Images: []domain.PersonImage{
		{ID: 5, ImageURL: "http://img/face.jpg"},
	},
}

var testFaceEmbeddings = []domain.FaceEmbedding{
	{ID: 1, PersonID: 1, Vector: []float64{0.1, 0.2, 0.3}, CreatedAt: "2026-06-27T17:00:00Z"},
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
		protected.GET("/me/device/persons/:personId/face-embeddings", personHandler.GetFaceEmbeddings)
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

func TestListPersons_ReturnsPersons(t *testing.T) {
	persons := []domain.Person{testPerson}

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons", testDeviceID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, persons))
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
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
	assert.Equal(t, "Juan Perez", first["name"])
	assert.Equal(t, "Cuidador", first["role"])
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

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreatePerson_ReturnsPerson(t *testing.T) {
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

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.CreatePersonRequest{
		Name:   "Juan Perez",
		Role:   "Cuidador",
		Images: []string{"base64img1"},
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

	assert.Equal(t, "Juan Perez", data["name"])
	assert.Equal(t, "Cuidador", data["role"])
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

	reqBody, _ := json.Marshal(domain.CreatePersonRequest{Name: "Juan", Role: "Cuidador"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/device/persons", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPerson_ReturnsPerson(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, testPerson))
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
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

	assert.Equal(t, "Juan Perez", data["name"])
	assert.Equal(t, float64(1), data["id"])
}

func TestGetPerson_DetectiveManagerError_ReturnsError(t *testing.T) {
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

func TestUpdatePerson_ReturnsUpdatedPerson(t *testing.T) {
	newName := "Juan Updated"
	updated := testPerson
	updated.Name = newName

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)

		var req domain.UpdatePersonRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, &newName, req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, updated))
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	reqBody, _ := json.Marshal(domain.UpdatePersonRequest{Name: &newName})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device/persons/"+testPersonID, bytes.NewBuffer(reqBody))
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

func TestUpdatePerson_InvalidBody_ReturnsBadRequest(t *testing.T) {
	router := setupPersonRouter("http://unused", "http://unused")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/device/persons/"+testPersonID, bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePerson_ReturnsNoContent(t *testing.T) {
	deleteCalled := false

	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		deleteCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.True(t, deleteCalled)
}

func TestDeletePerson_DetectiveManagerError_ReturnsError(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "person not found"})
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me/device/persons/"+testPersonID, nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
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

func TestGetFaceEmbeddings_ReturnsEmbeddings(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s/face-embeddings", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(encodeStdResponse(t, testFaceEmbeddings))
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons/"+testPersonID+"/face-embeddings", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&respBody))
	data := respBody["data"].([]interface{})

	require.Len(t, data, 1)
	first := data[0].(map[string]interface{})
	assert.Equal(t, float64(1), first["id"])
}

func TestGetFaceEmbeddings_DetectiveManagerError_ReturnsError(t *testing.T) {
	detectiveMux := http.NewServeMux()
	detectiveMux.HandleFunc(fmt.Sprintf("/v1/devices/%s/persons/%s/face-embeddings", testDeviceID, testPersonID), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"detail": "person not found"})
	})

	usersSrv := newUsersMock(t, usersMuxWithDevice(t))
	detectiveSrv := newDetectiveMock(t, detectiveMux)

	router := setupPersonRouter(usersSrv.URL, detectiveSrv.URL)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/device/persons/"+testPersonID+"/face-embeddings", nil)
	req.Header.Set("Authorization", buildBearerToken(t, testUserID))

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
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
