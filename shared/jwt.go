package shared

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var (
	serviceSecretKey []byte // Replace with your actual secret key
	secretKey        []byte
)

type ServiceCustomClaims struct {
	Service string `json:"service"`
	jwt.RegisteredClaims
}

func JWTServiceMiddleware(c *fiber.Ctx) error {
	jwtToken := GetTokenFromRequest(c)

	_, token, err := ValidateJWTForService(jwtToken)
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return c.Next()
}

func GenerateJWTForService(serviceName string) (string, error) {
	claims := ServiceCustomClaims{
		Service: serviceName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	secret := getServiceSecretKey()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateJWTForService(tokenString string) (*ServiceCustomClaims, *jwt.Token, error) {
	claims := &ServiceCustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return getServiceSecretKey(), nil
	})

	if err != nil {
		return nil, nil, err
	}

	return claims, token, nil
}

func GetTokenFromRequest(c *fiber.Ctx) string {
	jwtToken := c.Get("Authorization")

	return jwtToken
}

func getServiceSecretKey() []byte {
	if serviceSecretKey != nil {
		return serviceSecretKey
	}

	key := os.Getenv("SERVICE_JWT_SECRET_KEY")
	serviceSecretKey = []byte(key)

	return serviceSecretKey
}

func getSecretKey() []byte {
	if secretKey != nil {
		return secretKey
	}

	key := os.Getenv("USER_JWT_SECRET_KEY")
	secretKey = []byte(key)

	return secretKey
}
