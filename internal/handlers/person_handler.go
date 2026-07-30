package handlers

import (
	"errors"
	"net/http"

	"gateway-app/internal/domain"
	"gateway-app/internal/requestctx"
	"gateway-app/internal/service"
	"gateway-app/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type PersonHandler struct {
	service service.PersonService
}

func NewPersonHandler(svc service.PersonService) *PersonHandler {
	return &PersonHandler{service: svc}
}

// ListPersons godoc
// @Summary List persons for authenticated user's device
// @Description Returns all persons with their contact info registered on the authenticated user's linked device
// @Tags persons
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []domain.PersonWithContact
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/persons [get]
func (h *PersonHandler) ListPersons(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.ListPersons(c.Request.Context(), claims.User.Id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*[]domain.PersonWithContact]{Data: resp})
}

// CreatePerson godoc
// @Summary Create a person with contact on the authenticated user's device
// @Description Registers a new person with face images and contact information on the linked device
// @Tags persons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreatePersonWithContactRequest true "Person with contact payload"
// @Success 200 {object} domain.PersonWithContact
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/persons [post]
func (h *PersonHandler) CreatePerson(c *gin.Context) {
	var req domain.CreatePersonWithContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.CreatePerson(c.Request.Context(), claims.User.Id, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.PersonWithContact]{Data: resp})
}

// GetPerson godoc
// @Summary Get a person by ID from the authenticated user's device
// @Description Returns a single person with their contact info registered on the linked device
// @Tags persons
// @Produce json
// @Security BearerAuth
// @Param personId path string true "Person ID"
// @Success 200 {object} domain.PersonWithContact
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/persons/{personId} [get]
func (h *PersonHandler) GetPerson(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	personID := c.Param("personId")

	resp, err := h.service.GetPerson(c.Request.Context(), claims.User.Id, personID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.PersonWithContact]{Data: resp})
}

// UpdatePerson godoc
// @Summary Partially update a person and their contact on the authenticated user's device
// @Description Updates person and/or contact fields; all fields are optional
// @Tags persons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param personId path string true "Person ID"
// @Param request body domain.UpdatePersonWithContactRequest true "Fields to update"
// @Success 200 {object} domain.PersonWithContact
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/persons/{personId} [patch]
func (h *PersonHandler) UpdatePerson(c *gin.Context) {
	var req domain.UpdatePersonWithContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	personID := c.Param("personId")

	resp, err := h.service.UpdatePerson(c.Request.Context(), claims.User.Id, personID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.PersonWithContact]{Data: resp})
}

// DeletePerson godoc
// @Summary Delete a person and their contact from the authenticated user's device
// @Description Removes a registered person and their associated contact; NotFound from either service is tolerated
// @Tags persons
// @Security BearerAuth
// @Param personId path string true "Person ID"
// @Success 204
// @Failure 401 {object} apierror.APIError
// @Router /api/v1/me/device/persons/{personId} [delete]
func (h *PersonHandler) DeletePerson(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	personID := c.Param("personId")

	if err := h.service.DeletePerson(c.Request.Context(), claims.User.Id, personID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// AddPersonImages godoc
// @Summary Add face images to a person
// @Description Adds one or more base64-encoded face images to a registered person
// @Tags persons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param personId path string true "Person ID"
// @Param request body domain.AddPersonImagesRequest true "Images payload"
// @Success 200 {object} domain.Person
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/persons/{personId}/images [post]
func (h *PersonHandler) AddPersonImages(c *gin.Context) {
	var req domain.AddPersonImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	personID := c.Param("personId")

	resp, err := h.service.AddPersonImages(c.Request.Context(), claims.User.Id, personID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.Person]{Data: resp})
}

// DeletePersonImage godoc
// @Summary Delete a face image from a person
// @Description Removes a specific face image from a registered person
// @Tags persons
// @Security BearerAuth
// @Param personId path string true "Person ID"
// @Param imageId path string true "Image ID"
// @Success 204
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/persons/{personId}/images/{imageId} [delete]
func (h *PersonHandler) DeletePersonImage(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	personID := c.Param("personId")
	imageID := c.Param("imageId")

	if err := h.service.DeletePersonImage(c.Request.Context(), claims.User.Id, personID, imageID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *PersonHandler) handleServiceError(c *gin.Context, err error) {
	var detErr *domain.DetectorError
	if errors.As(err, &detErr) {
		apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
		return
	}
	apierror.Respond(c, apierror.InternalServerError("failed to process request", c.Request.URL.Path))
}

func (h *PersonHandler) handleJWTNotPresentError(c *gin.Context) {
	apierror.Respond(c, apierror.Unauthorized("missing claims", c.Request.URL.Path))
}
