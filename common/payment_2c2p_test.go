package common

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

// ensurePaymentSecret sets PAYMENT_2C2P_SECRET before tests touch the env.
// Unlike infra's sync.OnceValue caches, payment2C2PSecret reads on each
// call, so per-test t.Setenv is sufficient.
func ensurePaymentSecret(t *testing.T) {
	t.Helper()
	t.Setenv("PAYMENT_2C2P_SECRET", "test-2c2p-secret")
}

func TestSignJWTPayload_RoundTrip(t *testing.T) {
	ensurePaymentSecret(t)

	claims := jwt.MapClaims{
		"merchantID": "MID-001",
		"invoiceNo":  "INV-123",
		"amount":     499.5,
	}
	token, err := signJWTPayload(claims)
	if err != nil {
		t.Fatalf("signJWTPayload: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	parsed, err := parse2C2PJWTResponse(token)
	if err != nil {
		t.Fatalf("parse2C2PJWTResponse: %v", err)
	}
	if parsed["merchantID"] != "MID-001" {
		t.Fatalf("merchantID mismatch: %v", parsed["merchantID"])
	}
	if parsed["invoiceNo"] != "INV-123" {
		t.Fatalf("invoiceNo mismatch: %v", parsed["invoiceNo"])
	}
	if amount, _ := parsed["amount"].(float64); amount != 499.5 {
		t.Fatalf("amount mismatch: %v", parsed["amount"])
	}
}

func TestParse2C2PJWTResponse_WrongSigningMethod(t *testing.T) {
	ensurePaymentSecret(t)

	// Sign with RSA method (unsupported — must reject)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"x": "y"})
	signed, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	if _, err := parse2C2PJWTResponse(signed); err == nil {
		t.Fatal("expected reject for non-HMAC signing method")
	}
}

func TestParse2C2PJWTResponse_WrongSecret(t *testing.T) {
	ensurePaymentSecret(t)

	// Sign with different secret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"x": "y"})
	signed, _ := token.SignedString([]byte("WRONG-SECRET"))

	if _, err := parse2C2PJWTResponse(signed); err == nil {
		t.Fatal("expected reject for wrong-secret signature")
	}
}

func TestParse2C2PJWTResponse_Malformed(t *testing.T) {
	ensurePaymentSecret(t)

	if _, err := parse2C2PJWTResponse("not.a.jwt"); err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestDecodePaymentResponse_ExtractsClaims(t *testing.T) {
	ensurePaymentSecret(t)

	claims := jwt.MapClaims{
		"webPaymentUrl": "https://pay.example.com/checkout/abc",
		"paymentToken":  "tok-123",
		"respCode":      "0000",
		"respDesc":      "Success",
	}
	token, err := signJWTPayload(claims)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	resp, err := decodePaymentResponse(token)
	if err != nil {
		t.Fatalf("decodePaymentResponse: %v", err)
	}
	if resp.WebPaymentUrl != "https://pay.example.com/checkout/abc" {
		t.Errorf("WebPaymentUrl: %q", resp.WebPaymentUrl)
	}
	if resp.PaymentToken != "tok-123" {
		t.Errorf("PaymentToken: %q", resp.PaymentToken)
	}
	if resp.RespCode != "0000" {
		t.Errorf("RespCode: %q", resp.RespCode)
	}
	if resp.RespDesc != "Success" {
		t.Errorf("RespDesc: %q", resp.RespDesc)
	}
}

func TestDecodePaymentResponse_MissingClaimsReturnEmpty(t *testing.T) {
	ensurePaymentSecret(t)

	// Sign a token with NO payment-response claims
	token, err := signJWTPayload(jwt.MapClaims{"unrelated": "x"})
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	resp, err := decodePaymentResponse(token)
	if err != nil {
		t.Fatalf("decodePaymentResponse: %v", err)
	}
	if resp.WebPaymentUrl != "" || resp.PaymentToken != "" || resp.RespCode != "" || resp.RespDesc != "" {
		t.Fatalf("expected all empty for missing claims, got %+v", resp)
	}
}

func TestGetStringClaim(t *testing.T) {
	cases := []struct {
		name   string
		claims jwt.MapClaims
		key    string
		want   string
	}{
		{"present", jwt.MapClaims{"k": "v"}, "k", "v"},
		{"missing", jwt.MapClaims{}, "k", ""},
		{"wrong type (int)", jwt.MapClaims{"k": 42}, "k", ""},
		{"wrong type (nil)", jwt.MapClaims{"k": nil}, "k", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getStringClaim(tc.claims, tc.key); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestGetFloat64Claim(t *testing.T) {
	cases := []struct {
		name   string
		claims jwt.MapClaims
		key    string
		want   float64
	}{
		{"present", jwt.MapClaims{"amount": 99.5}, "amount", 99.5},
		{"zero", jwt.MapClaims{"amount": 0.0}, "amount", 0},
		{"missing", jwt.MapClaims{}, "amount", 0},
		{"wrong type (string)", jwt.MapClaims{"amount": "99.5"}, "amount", 0},
		{"wrong type (int)", jwt.MapClaims{"amount": 99}, "amount", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getFloat64Claim(tc.claims, tc.key); got != tc.want {
				t.Fatalf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestPayment2C2PSecret_ReadsEnvLive(t *testing.T) {
	t.Setenv("PAYMENT_2C2P_SECRET", "secret-v1")
	if string(payment2C2PSecret()) != "secret-v1" {
		t.Fatalf("did not read v1")
	}
	t.Setenv("PAYMENT_2C2P_SECRET", "secret-v2")
	if string(payment2C2PSecret()) != "secret-v2" {
		t.Fatal("did not re-read after env change — would break secret rotation")
	}
}
