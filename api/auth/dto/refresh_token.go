package dto

import "github.com/go-playground/validator/v10"

// =======================================
// ||         Refresh Token Request      ||
// =======================================

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func NewRefreshTokenRequest() *RefreshTokenRequest {
	return &RefreshTokenRequest{}
}

func (r *RefreshTokenRequest) GetValue() *RefreshTokenRequest {
	return r
}

func (r *RefreshTokenRequest) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, "Refresh token is required")
		default:
			msgs = append(msgs, "Invalid refresh token")
		}
	}
	return msgs, nil
}

// =======================================
// ||         Refresh Token Response    ||
// =======================================

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshTokenResponse(accessToken, refreshToken string) *RefreshTokenResponse {
	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

func (r *RefreshTokenResponse) GetValue() *RefreshTokenResponse {
	return r
}

func (r *RefreshTokenResponse) ValidateErrors(errs validator.ValidationErrors) ([]string, error) {
	var msgs []string
	for _, err := range errs {
		switch err.Tag() {
		case "required":
			msgs = append(msgs, "Access token and refresh token are required")
		default:
			msgs = append(msgs, "Invalid access token or refresh token")
		}
	}
	return msgs, nil
}
