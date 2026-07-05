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

type ContactHandler struct {
	service service.ContactService
}

func NewContactHandler(svc service.ContactService) *ContactHandler {
	return &ContactHandler{service: svc}
}

// ListContacts godoc
// @Summary List contacts for authenticated user's device
// @Description Returns all contacts registered on the authenticated user's linked device
// @Tags contacts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []domain.Contact
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/contacts [get]
func (h *ContactHandler) ListContacts(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.ListContacts(c.Request.Context(), claims.User.Id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*[]domain.Contact]{Data: resp})
}

// CreateContact godoc
// @Summary Create a contact on the authenticated user's device
// @Description Adds a secondary contact to the linked device
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateContactRequest true "Contact payload"
// @Success 200 {object} domain.Contact
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/contacts [post]
func (h *ContactHandler) CreateContact(c *gin.Context) {
	var req domain.CreateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.CreateContact(c.Request.Context(), claims.User.Id, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.Contact]{Data: resp})
}

// UpdateContact godoc
// @Summary Partially update a contact on the authenticated user's device
// @Description Updates contact information for a registered contact
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param contactId path string true "Contact ID"
// @Param request body domain.UpdateContactRequest true "Fields to update"
// @Success 200 {object} domain.Contact
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/contacts/{contactId} [patch]
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	var req domain.UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	contactID := c.Param("contactId")

	resp, err := h.service.UpdateContact(c.Request.Context(), claims.User.Id, contactID, req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.Contact]{Data: resp})
}

// DeleteContact godoc
// @Summary Delete a contact from the authenticated user's device
// @Description Removes a registered contact
// @Tags contacts
// @Security BearerAuth
// @Param contactId path string true "Contact ID"
// @Success 204
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device/contacts/{contactId} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	contactID := c.Param("contactId")

	if err := h.service.DeleteContact(c.Request.Context(), claims.User.Id, contactID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContactHandler) handleServiceError(c *gin.Context, err error) {
	var detErr *domain.DetectorError
	if errors.As(err, &detErr) {
		apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
		return
	}
	apierror.Respond(c, apierror.InternalServerError("failed to process request", c.Request.URL.Path))
}

func (h *ContactHandler) handleJWTNotPresentError(c *gin.Context) {
	apierror.Respond(c, apierror.Unauthorized("missing claims", c.Request.URL.Path))
}
