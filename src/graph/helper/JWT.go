package helper

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	email                string
	name                 string
	google_id            string
	ids                  string
	jwt.RegisteredClaims // Thêm các trường chuẩn như exp, iat
}

type ContextKey string

const Auth ContextKey = "xxxyyyzzzkkk"

// ParseJWT parse token string thành claims
func ParseJWT(tokenString string, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra thuật toán ký
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("phương thức ký không hợp lệ: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Ép kiểu sang MapClaims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		return claims, nil
	}
	return nil, errors.New("token không hợp lệ")
}

// ExtractBearerToken trích xuất token từ Authorization header
// Header format: "Bearer <token>"
// Trả về token string hoặc error nếu format không đúng
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization format, expected 'Bearer <token>'")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// ValidateAndParseClaims validate và parse token từ Authorization header
// Kết hợp ExtractBearerToken và ParseJWT thành một hàm tiện lợi
func ValidateAndParseClaims(authHeader string, secret string) (jwt.MapClaims, error) {
	token, err := ExtractBearerToken(authHeader)
	if err != nil {
		return nil, err
	}

	claims, err := ParseJWT(token, secret)

	if err != nil {
		return nil, err
	}

	return claims, nil
}
