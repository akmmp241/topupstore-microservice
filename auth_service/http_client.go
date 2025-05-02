package main

import (
	"fmt"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/gofiber/fiber/v2"
	"os"
	"time"
)

var (
	cachedToken     string
	tokenExpiryTime time.Time
)

type HttpClientRes struct {
	StatusCode int
	Body       []byte
	Errs       []error
}

func getServiceToken(serviceName string) (string, error) {
	if cachedToken != "" && time.Now().Before(tokenExpiryTime) {
		return cachedToken, nil
	}

	// Token expired or not yet created
	token, err := shared.GenerateJWTForService(serviceName)
	if err != nil {
		return "", err
	}

	// Set new expiry manually (you can extract from JWT too)
	cachedToken = token
	tokenExpiryTime = time.Now().Add(14 * time.Minute) // a little buffer

	return token, nil
}

func CallUserService(url string, method string, body interface{}) (*HttpClientRes, error) {
	userServiceHost := os.Getenv("USER_SERVICE_HOST")
	userServicePort := os.Getenv("USER_SERVICE_PORT")
	userServiceURL := fmt.Sprintf("http://%s:%s", userServiceHost, userServicePort)

	jwtForService, err := getServiceToken("auth-service")
	if err != nil {
		return nil, err
	}

	url = fmt.Sprintf("%s/api/%s", userServiceURL, url)

	var agent *fiber.Agent

	switch method {
	case fiber.MethodGet:
		agent = fiber.Get(url)
	case fiber.MethodPost:
		agent = fiber.Post(url)
	case fiber.MethodPut:
		agent = fiber.Put(url)
	case fiber.MethodDelete:
		agent = fiber.Delete(url)
	case fiber.MethodPatch:
		agent = fiber.Patch(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if body != nil {
		agent.JSON(body)
	}

	agent.Add("Authorization", fmt.Sprintf("Bearer %s", jwtForService))
	statusCode, respBody, errs := agent.Bytes()

	return &HttpClientRes{
		StatusCode: statusCode,
		Body:       respBody,
		Errs:       errs,
	}, nil
}
