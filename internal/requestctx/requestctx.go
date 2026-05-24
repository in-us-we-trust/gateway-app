package requestctx

import (
	"context"

	"gateway-app/internal/domain"
)

type ctxKey string

const claimsKey ctxKey = "jwt_claims"

func WithClaims(
	ctx context.Context,
	claims domain.JWTClaims,
) context.Context {

	return context.WithValue(
		ctx,
		claimsKey,
		claims,
	)
}

func Claims(
	ctx context.Context,
) (domain.JWTClaims, bool) {

	claims, ok := ctx.Value(claimsKey).(domain.JWTClaims)

	if !ok {
		return domain.JWTClaims{}, false
	}

	return claims, true
}
