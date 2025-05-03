package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type GoogleToken struct {
	Sub           string `json:"sub"`
	Iss           string `json:"iss"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
}

func DecodeGoogleJWTToken(token string) (*GoogleToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding token payload: %v", err)
	}

	var googleToken GoogleToken
	err = json.Unmarshal(payload, &googleToken)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling token payload: %v", err)
	}

	return &googleToken, nil
}
