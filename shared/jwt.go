package shared

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"os"
	"strings"
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
	jwtToken, err := GetTokenFromRequest(c)
	if err != nil {
		slog.Error("Error getting token from request", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	_, token, err := ValidateJWTForService(jwtToken)
	if err != nil || !token.Valid {
		slog.Error("Error validating token", "err", err)
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

func JWTUserMiddleware(c *fiber.Ctx) error {
	jwtToken, err := GetTokenFromRequest(c)
	if err != nil {
		slog.Error("Error getting token from request", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	token, err := ValidateJWTForUser(jwtToken)
	if err != nil || !token.Valid {
		slog.Error("Error validating token", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return c.Next()
}

func GenerateJWTForUser(userID string, expiry time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    "topupstore-microservice",
		ExpiresAt: jwt.NewNumericDate(expiry),
		NotBefore: jwt.NewNumericDate(time.Now()),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	secret := getSecretKey()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateJWTForUser(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the secret key for validation
		return getSecretKey(), nil
	})
}

func GetUserIdFromToken(c *fiber.Ctx) (string, error) {
	token, err := GetTokenFromRequest(c)
	if err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	claims, err := ValidateJWTForUser(token)
	if err != nil {
		return "", fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return claims.Claims.(jwt.RegisteredClaims).Subject, nil
}

func GetTokenFromRequest(c *fiber.Ctx) (string, error) {
	// Extract JWT from the Authorization header
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("missing or invalid Authorization header")
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	return token, nil
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
