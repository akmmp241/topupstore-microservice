package shared

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var (
	secretKey []byte // Replace with your actual secret key
)

type CustomClaims struct {
	Service string `json:"service"`
	jwt.RegisteredClaims
}

func GenerateJWTForService(serviceName string) (string, error) {
	claims := CustomClaims{
		Service: serviceName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	secret := getSecretKey()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func getSecretKey() []byte {
	if secretKey != nil {
		return secretKey
	}

	key := os.Getenv("SERVICE_JWT_SECRET_KEY")
	secretKey = []byte(key)

	return secretKey
}
