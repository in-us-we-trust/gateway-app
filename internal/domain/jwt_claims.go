package domain

// Hay mas campos, pero me interesa el userId
type UserJwt struct {
	Id string `json:"id"`
}

type JWTClaims struct {
	User UserJwt `json:"user"`
}
