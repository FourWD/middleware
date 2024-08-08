package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

type FirebaseIdentities struct {
	Google   []string `json:"google.com"`
	FaceBook []string `json:"facebook.com"`
	Apple    []string `json:"apple"`

	Email []string `json:"email"`
}

type Firebase struct {
	Identities     FirebaseIdentities `json:"identities"`
	SignInProvider string             `json:"sign_in_provider"`
}

type JWTClaimsDeCode struct {
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Email   string `json:"email"`
	UserID  string `json:"user_id"`

	EmailVerified bool     `json:"email_verified"`
	Firebase      Firebase `json:"firebase"`
	jwt.RegisteredClaims
}

func DecodeFirebaseToken(tokenString string) (*JWTClaimsDeCode, error) {

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &JWTClaims{})

	if err != nil {
		fmt.Println("Error parsing token:", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaimsDeCode); ok {
		signInProvider := claims.Firebase.SignInProvider
		fmt.Println("Sign-in Provider:", signInProvider)

		// Print the claims
		jsonString, _ := json.MarshalIndent(claims, "", "  ")
		fmt.Println(string(jsonString))
		return claims, nil
	} else {
		fmt.Println("Invalid JWT Token")
	}

	return nil, errors.New("error decode")
}
