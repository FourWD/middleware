package infra

import "github.com/gofiber/fiber/v3"

// GetSession returns the JWT claims attached to the request by
// AuthenticationMiddleware, or nil if the middleware did not populate them
// (logs SESSION_INVALID_SIGNATURE in that case).
func GetSession(c fiber.Ctx) *JWTClaims {
	claims := fiber.Locals[*JWTClaims](c, "user")
	if claims == nil {
		logInvalidSession(c)
	}
	return claims
}

// GetSessionUserID returns the user_id for the current request. Falls back
// to extracting "user_id" from the raw JWT claim if AuthenticationMiddleware
// did not attach a session — covers code paths that read user_id before
// middleware runs.
func GetSessionUserID(c fiber.Ctx) string {
	session := GetSession(c)
	if session == nil {
		userID, _ := EncodedJwtToken(c, "user_id")
		return userID
	}
	return session.UserID
}

// logInvalidSession emits a security warning when the JWT middleware did not
// attach a claims object. The raw Authorization header is intentionally NOT
// logged — it carries the bearer token (PII deny-list).
func logInvalidSession(c fiber.Ctx) {
	AppLog.WarnCtx(c.Context(), "SESSION_INVALID_SIGNATURE", nil,
		WithComponent(ComponentAuth),
		WithOperation("read_session"),
		WithLogKind(LogKindSecurity))
}
