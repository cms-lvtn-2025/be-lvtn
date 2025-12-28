package model

// GoogleLoginResponse response for Google login
type GoogleLoginResponse struct {
	AuthURL string `json:"auth_url"`
}

// GoogleCallbackRequest request for Google callback
type GoogleCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state"`
	Role  string `json:"role"`
}

// RefreshTokenRequest request for refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest request for logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse response containing token pair
type TokenResponse struct {
	GoogleUser   interface{} `json:"google_user,omitempty"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	TokenType    string      `json:"token_type"`
}
