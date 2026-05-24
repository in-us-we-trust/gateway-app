package middleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"gateway-app/internal/domain"
	"gateway-app/internal/requestctx"
	"gateway-app/pkg/apierror"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" ||
			!strings.HasPrefix(authHeader, "Bearer ") {

			apierror.Respond(
				c,
				apierror.Unauthorized(
					"authorization header with bearer token is required",
					c.Request.URL.Path,
				),
			)

			c.Abort()
			return
		}

		token := strings.TrimSpace(
			strings.TrimPrefix(authHeader, "Bearer "),
		)

		claims, err := extractClaimsFromJWT(token)
		if err != nil {
			apierror.Respond(
				c,
				apierror.Unauthorized(
					"invalid JWT payload: user_id is required",
					c.Request.URL.Path,
				),
			)

			c.Abort()
			return
		}

		ctx := requestctx.WithClaims(
			c.Request.Context(),
			*claims,
		)

		c.Request = c.Request.WithContext(ctx)

		// TODO: validate JWT token signature and claims against trusted issuer
		log.Printf(
			"[security] JWT validation performed for %s %s",
			c.Request.Method,
			c.Request.URL.Path,
		)

		c.Next()
	}
}

func extractClaimsFromJWT(
	token string,
) (*domain.JWTClaims, error) {

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidJWT
	}

	payloadBytes, err := base64.RawURLEncoding.
		DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidJWT
	}

	var claims domain.JWTClaims

	if err := json.Unmarshal(
		payloadBytes,
		&claims,
	); err != nil {

		return nil, errInvalidJWT
	}

	claims.UserID = strings.TrimSpace(
		claims.UserID,
	)

	claims.DeviceID = strings.TrimSpace(
		claims.DeviceID,
	)

	if claims.UserID == "" {
		return nil, errInvalidJWT
	}

	return &claims, nil
}

var errInvalidJWT = errors.New("invalid jwt payload")
