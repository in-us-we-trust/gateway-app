package handlers

import (
	"errors"
	"gateway-app/internal/domain"
	"gateway-app/internal/service"
	"gateway-app/pkg/apierror"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Register godoc
// @Summary Register a new user
// @Description Creates auth credentials and user profile
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Register payload"
// @Success 201 {object} domain.UserResponse
// @Failure 400 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, apierror.BadRequest(err.Error(), c.Request.URL.Path))
		return
	}

	resp, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		var detErr *domain.DetectorError
		if errors.As(err, &detErr) {
			apierror.Respond(c, apierror.FromDetectorError(detErr.DetectorDetail, c.Request.URL.Path, detErr.Code, detErr.StatusCode))
			return
		}
		apierror.Respond(c, apierror.InternalServerError("failed to register user", c.Request.URL.Path))
		return
	}

	c.JSON(http.StatusCreated, resp)
}
