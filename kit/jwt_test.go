package kit

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestDecodeJWT_RejectsMalleatedSignature(t *testing.T) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "1", "exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte("s"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeJWT(token, "s"); err != nil {
		t.Fatalf("canonical token must decode: %v", err)
	}
	// HS256 signatures leave 2 unused low bits in the last char.
	idx := strings.IndexByte(alphabet, token[len(token)-1])
	alt := token[:len(token)-1] + string(alphabet[idx^1])
	if _, err := DecodeJWT(alt, "s"); err == nil {
		t.Fatal("malleated signature must be rejected")
	}
}
