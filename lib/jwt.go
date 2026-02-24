package lib

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func GenerateTokenJwt(payload map[string]interface{}) (string, error) {
	token:= jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))

	tokenString, err := token.SignedString([]byte("secret"))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyTokenJwt verifies the JWT token string and returns the claims
func VerifyTokenJwt(tokenString string) (jwt.MapClaims, error) {
	godotenv.Load()
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		// The secret key should match the one used during generation ("secret")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}