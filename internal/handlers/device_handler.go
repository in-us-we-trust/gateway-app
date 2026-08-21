package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"gateway-app/internal/domain"
	"gateway-app/internal/requestctx"
	"gateway-app/internal/service"
	"gateway-app/pkg/apierror"

	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	service service.DeviceService
}

func NewDeviceHandler(svc service.DeviceService) *DeviceHandler {
	return &DeviceHandler{service: svc}
}

// LinkMyDevice godoc
// @Summary Link device to authenticated user
// @Description Associates a device with the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.LinkDeviceRequest true "Device payload"
// @Success 200 {object} domain.DeviceResponseDto
// // @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device [post]
func (h *DeviceHandler) LinkMyDevice(c *gin.Context) {
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

	resp, err := h.service.LinkDevice(c.Request.Context(), claims.User.Id, req)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.DeviceResponseDto]{Data: resp})
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
func (h *DeviceHandler) UnlinkMyDevice(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	err := h.service.UnlinkDevice(c.Request.Context(), claims.User.Id)
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

// UpdateMyDevice godoc
// @Summary Update my linked device
// @Description Partially updates the authenticated user's linked device. Only city, detectionThreshold and location can be modified.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateDeviceRequest true "Fields to update"
// @Success 200 {object} domain.DeviceResponseDto
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device [patch]
func (h *DeviceHandler) UpdateMyDevice(c *gin.Context) {
	var req domain.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.UpdateMyDevice(c.Request.Context(), claims.User.Id, req)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to update device", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.DeviceResponseDto]{Data: resp})
}

// GetMyDevice godoc
// @Summary Get my linked device
// @Description Returns the authenticated user's linked device
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.DeviceResponseDto
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/me/device [get]
func (h *DeviceHandler) GetMyDevice(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.GetMyDevice(c.Request.Context(), claims.User.Id)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.DeviceResponseDto]{Data: resp})
}

// GetMyDeviceEvents godoc
// @Summary Get authenticated user's device events
// @Description Returns all events associated with the authenticated user's linked device
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.TracksResult
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/me/device/events [get]
func (h *DeviceHandler) GetMyDeviceEvents(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.GetMyDeviceEvents(c.Request.Context(), claims.User.Id)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.TracksResult]{Data: resp})
}

// GetMyDeviceEventDetail godoc
// @Summary Get authenticated user's device event detail
// @Description Returns the detail of a single track (visit) of the user's device
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param trackId path string true "Track ID"
// @Success 200 {object} domain.TrackDetail
// @Failure 401 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/me/device/events/{trackId} [get]
func (h *DeviceHandler) GetMyDeviceEventDetail(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	trackID := c.Param("trackId")

	resp, err := h.service.GetMyDeviceEventDetail(c.Request.Context(), claims.User.Id, trackID)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to process user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, domain.StandardResponse[*domain.TrackDetail]{Data: resp})
}

func (h *DeviceHandler) handleJWTNotPresentError(
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

// RegisterPushToken godoc
// @Summary Register push notification token
// @Description Registers or updates the FCM push token for the authenticated user's device
// @Tags users
// @Accept json
// @Security BearerAuth
// @Param request body domain.RegisterPushTokenRequest true "Push token payload"
// @Success 200
// @Failure 400 {object} apierror.APIError
// @Failure 401 {object} apierror.APIError
// @Router /api/v1/me/push-token [post]
func (h *DeviceHandler) RegisterPushToken(c *gin.Context) {
	var req domain.RegisterPushTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	resp, err := h.service.RegisterPushToken(c.Request.Context(), claims.User.Id, req)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to register push token", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeletePushToken godoc
// @Summary Delete push notification token
// @Description Removes the push token from an app device without deleting the device
// @Tags users
// @Security BearerAuth
// @Param appDeviceId path int true "App Device ID"
// @Success 204
// @Failure 401 {object} apierror.APIError
// @Failure 403 {object} apierror.APIError
// @Router /api/v1/me/push-token/{appDeviceId} [delete]
func (h *DeviceHandler) DeletePushToken(c *gin.Context) {
	claims, ok := requestctx.Claims(c.Request.Context())
	if !ok {
		h.handleJWTNotPresentError(c)
		return
	}

	appDeviceID, err := strconv.ParseInt(c.Param("appDeviceId"), 10, 64)
	if err != nil {
		apierror.Respond(c, apierror.BadRequest("invalid appDeviceId", c.Request.URL.Path))
		return
	}

	if err := h.service.DeletePushToken(c.Request.Context(), claims.User.Id, appDeviceID); err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to delete push token", c.Request.URL.Path))
		return
	}

	c.Status(http.StatusNoContent)
}
