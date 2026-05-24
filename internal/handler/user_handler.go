package handler

import (
	"errors"
	"net/http"

	"gateway-app/internal/domain"
	"gateway-app/internal/requestctx"
	"gateway-app/internal/service"
	"gateway-app/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// LinkMyDevice godoc
// @Summary Link device to authenticated user
// @Description Associates a device with the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.LinkDeviceRequest true "Device payload"
// @Success 200 {object} domain.DeviceResponse
// // @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device [post]
func (h *UserHandler) LinkMyDevice(c *gin.Context) {
	var req domain.LinkDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.LinkDevice(c.Request.Context(), claims.UserID, req)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UnlinkMyDevice godoc
// @Summary Unlink device from authenticated user
// @Description Removes the linked device from the authenticated user
// @Tags users
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device [delete]
func (h *UserHandler) UnlinkMyDevice(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	err := h.service.UnlinkDevice(c.Request.Context(), claims.UserID)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetMyDevice godoc
// @Summary Get my linked device
// @Description Returns the authenticated user's linked device
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DeviceResponse
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device [get]
func (h *UserHandler) GetMyDevice(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.GetMyDevice(c.Request.Context(), claims.UserID)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMyDeviceEvents godoc
// @Summary Get authenticated user's device events
// @Description Returns all events associated with the authenticated user's linked device
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.EventsResponse
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/me/device/events [get]
func (h *UserHandler) GetMyDeviceEvents(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.GetMyDeviceEvents(c.Request.Context(), claims.UserID)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) handleJWTNotPresentError(
	c *gin.Context,
) {
	apierror.Respond(
		c,
		apierror.Unauthorized(
			"missing claims",
			c.Request.URL.Path,
		),
	)
}
