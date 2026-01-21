package lib

import "github.com/golang-jwt/jwt/v5"

func GenerateTokenJwt(payload map[string]interface{}) (string, error) {
	token:= jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))

	tokenString, err := token.SignedString([]byte("secret"))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}