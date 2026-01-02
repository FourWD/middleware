package common

import (
	"context"
	"errors"
	"time"

	"firebase.google.com/go/v4/auth"
)

type FirebaseIdentities struct {
	Google   []string `json:"google.com"`
	FaceBook []string `json:"facebook.com"`
	Apple    []string `json:"apple"`
	Email    []string `json:"email"`
}

type Firebase struct {
	Identities     FirebaseIdentities `json:"identities"`
	SignInProvider string             `json:"sign_in_provider"`
}

type JWTClaimsDeCode struct {
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Email         string   `json:"email"`
	UserID        string   `json:"user_id"`
	EmailVerified bool     `json:"email_verified"`
	Firebase      Firebase `json:"firebase"`
	Issuer        string   `json:"iss"`
	Subject       string   `json:"sub"`
	Audience      string   `json:"aud"`
	ExpiresAt     int64    `json:"exp"`
	IssuedAt      int64    `json:"iat"`
}

// DecodeFirebaseToken verifies and decodes a Firebase ID token.
// This function properly validates the token signature using Firebase Admin SDK.
func DecodeFirebaseToken(tokenString string) (*JWTClaimsDeCode, error) {
	if AuthClient == nil {
		return nil, errors.New("firebase auth client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// VerifyIDToken validates the signature, expiration, issuer, and audience
	token, err := AuthClient.VerifyIDToken(ctx, tokenString)
	if err != nil {
		AppLog.Error("Error verifying Firebase token: " + err.Error())
		return nil, err
	}

	// Map Firebase token to JWTClaimsDeCode
	claims := &JWTClaimsDeCode{
		UserID:    token.UID,
		Issuer:    token.Issuer,
		Subject:   token.Subject,
		Audience:  token.Audience,
		ExpiresAt: token.Expires,
		IssuedAt:  token.IssuedAt,
	}

	// Extract additional claims from token
	if name, ok := token.Claims["name"].(string); ok {
		claims.Name = name
	}
	if picture, ok := token.Claims["picture"].(string); ok {
		claims.Picture = picture
	}
	if email, ok := token.Claims["email"].(string); ok {
		claims.Email = email
	}
	if emailVerified, ok := token.Claims["email_verified"].(bool); ok {
		claims.EmailVerified = emailVerified
	}

	// Extract Firebase-specific claims
	if firebaseClaim, ok := token.Claims["firebase"].(map[string]interface{}); ok {
		if signInProvider, ok := firebaseClaim["sign_in_provider"].(string); ok {
			claims.Firebase.SignInProvider = signInProvider
		}
		if identities, ok := firebaseClaim["identities"].(map[string]interface{}); ok {
			claims.Firebase.Identities = extractIdentities(identities)
		}
	}

	return claims, nil
}

// VerifyFirebaseToken verifies a Firebase ID token and returns the raw auth.Token.
// Use this when you need access to all token claims.
func VerifyFirebaseToken(tokenString string) (*auth.Token, error) {
	if AuthClient == nil {
		return nil, errors.New("firebase auth client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := AuthClient.VerifyIDToken(ctx, tokenString)
	if err != nil {
		AppLog.Error("Error verifying Firebase token: " + err.Error())
		return nil, err
	}

	return token, nil
}

func extractIdentities(identities map[string]interface{}) FirebaseIdentities {
	var fi FirebaseIdentities

	if google, ok := identities["google.com"].([]interface{}); ok {
		fi.Google = interfaceSliceToStringSlice(google)
	}
	if facebook, ok := identities["facebook.com"].([]interface{}); ok {
		fi.FaceBook = interfaceSliceToStringSlice(facebook)
	}
	if apple, ok := identities["apple"].([]interface{}); ok {
		fi.Apple = interfaceSliceToStringSlice(apple)
	}
	if email, ok := identities["email"].([]interface{}); ok {
		fi.Email = interfaceSliceToStringSlice(email)
	}

	return fi
}

func interfaceSliceToStringSlice(slice []interface{}) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
