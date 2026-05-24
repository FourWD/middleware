package infra

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// GetNotiToken returns the device notification token. Priority order:
//  1. JWT claim "noti_token" from Authorization header
//  2. session.Remark["noti_token"] (set by login)
//
// The token value is never logged (PII deny-list).
func GetNotiToken(c fiber.Ctx) (string, error) {
	notiToken, _ := EncodedJwtToken(c, "noti_token")

	if notiToken == "" {
		session := GetSession(c)
		if session == nil {
			return "", nil
		}

		if value, ok := session.Remark["noti_token"]; ok {
			AppLog.EventCtx(c.Context(), "NOTI_TOKEN_FROM_SESSION", nil,
				WithComponent(ComponentAuth),
				WithOperation("read_noti_token"),
				WithLogKind(LogKindDiagnostic))
			return value, nil
		}

		return "", errors.New("notiToken is nil")
	}

	return notiToken, nil
}
