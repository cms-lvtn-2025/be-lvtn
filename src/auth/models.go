package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session model lưu thông tin phiên đăng nhập trong MongoDB
type Session struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Ids          string             `bson:"ids" json:"ids"`
	Role         string             `bson:"role" json:"role"`
	UserID       string             `bson:"user_id" json:"user_id"` // User ID từ service user
	Email        string             `bson:"email" json:"email"`     // Email để reference
	RefreshToken string             `bson:"refresh_token" json:"refresh_token"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

// GoogleUserInfo thông tin user từ Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// TokenPair chứa access token và refresh token
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	TokenType    string `json:"token_type"`
}

// RefreshTokenClaims lưu trong Redis
type RefreshTokenClaims struct {
	UserID    string    `json:"user_id"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
