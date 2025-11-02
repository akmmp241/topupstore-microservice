package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net"
	"os"
	"strings"
	"time"

	ppb "github.com/akmmp241/topupstore-microservice/payment-proto/v1"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	ListenAddr  string
	DB          *sql.DB
	Server      *grpc.Server
	NetListener net.Listener
	ppb.UnimplementedPaymentServiceServer
}

func (s *GrpcServer) CreatePayment(ctx context.Context, req *ppb.CreatePaymentReq) (*ppb.CreatePaymentRes, error) {
	xenditRequestBody := XenditRequestBody{
		Currency:      "IDR",
		Country:       "ID",
		Type:          "PAY",
		CaptureMethod: "AUTOMATIC",
		RequestAmount: int(math.Ceil(float64(req.Amount))),
		ReferenceId:   req.ReferenceId,
	}

	displayName := strings.Replace(req.BuyerEmail, "@", "..", 1)
	xenditRequestBody.ChannelProperties = ChannelProperties{
		DisplayName:      displayName,
		ExpiresAt:        time.Now().Add(time.Hour),
		SuccessReturnUrl: "https://www.xendit.co/success",
		FailureReturnUrl: "https://www.xendit.co/failure",
		CancelReturnUrl:  "https://www.xendit.co/cancel",
	}

	// Check if the payment method is valid
	for _, channel := range EwalletChannelCodes {
		if req.ChannelCode != channel {
			continue
		}

		// implementation for creating ewallet payment
		xenditRequestBody.ChannelCode = channel
	}

	for _, channel := range VirtualAccountChannelCodes {
		if req.ChannelCode != channel {
			continue
		}

		// implementation for creating virtual account payment
		xenditRequestBody.ChannelCode = channel
	}

	for _, channel := range QrisChannelCode {
		if req.ChannelCode != channel {
			continue
		}

		// implementation for creating qris payment
		xenditRequestBody.ChannelCode = channel
	}

	// check if channel code is valid
	if xenditRequestBody.ChannelCode == "" {
		slog.Error("Channel code is not valid", "channel_code", req.ChannelCode)
		return nil, status.Error(codes.InvalidArgument, "Channel code is not valid")
	}

	xenditApiKey := os.Getenv("XENDIT_API_KEY") + ":"
	xenditApiKeyBase64 := base64.StdEncoding.EncodeToString([]byte(xenditApiKey))
	xenditHost := os.Getenv("XENDIT_API_URL")
	paymentReqUrl := fmt.Sprintf("%s/v3/payment_requests", xenditHost)

	agent := fiber.Post(paymentReqUrl).Timeout(15*time.Second).
		Add("Authorization", fmt.Sprintf("Basic %s", xenditApiKeyBase64)).
		Add("api-version", "2024-11-11").
		ContentType(fiber.MIMEApplicationJSON).JSON(xenditRequestBody)

	statusCode, respByte, errs := agent.Bytes()

	if len(errs) > 0 {
		slog.Error(
			"Error occurred while calling xendit payment request api",
			"errs",
			errs,
			"resp",
			string(respByte),
		)
		return nil, errs[0]
	}

	if statusCode >= 400 && statusCode < 500 {
		var errMsg XenditErrMsg
		err := json.Unmarshal(respByte, &errMsg)
		if err != nil {
			return nil, err
		}
		slog.Error(
			"xendit payment request api returned 4xx status code",
			"code", statusCode,
			"resp", string(respByte),
			"err", err,
		)

		if statusCode == 400 {
			return nil, status.Error(codes.InvalidArgument, errMsg.Message)
		}

		if statusCode == 403 {
			return nil, status.Error(codes.PermissionDenied, errMsg.Message)
		}

		if statusCode == 409 {
			return nil, status.Error(codes.AlreadyExists, errMsg.Message)
		}
	}

	var paymentRequestResponse XenditPaymentRequestResponse
	err := json.Unmarshal(respByte, &paymentRequestResponse)
	if err != nil {
		slog.Error(
			"Error occurred while unmarshalling xendit payment request api response",
			"err",
			err,
		)
		return nil, err
	}

	res := &ppb.CreatePaymentRes{
		XenditPaymentId: paymentRequestResponse.PaymentRequestId,
		Status:          paymentRequestResponse.Status,
		FailureCode:     paymentRequestResponse.FailureCode,
	}

	return res, nil
}

func (s *GrpcServer) GetPaymentById(ctx context.Context, req *ppb.GetPaymentByIdReq) (*ppb.GetPaymentByIdRes, error) {
	slog.Debug("Getting payment by id", "req", req)
	apiKey, baseUrl := getXenditCredentials()

	agent := fiber.Get(baseUrl+req.GetPaymentId()).Timeout(15*time.Second).
		Add("Authorization", fmt.Sprintf("Basic %s", apiKey)).
		Add("api-version", "2024-11-11")

	statusCode, respByte, errs := agent.Bytes()

	if len(errs) > 0 {
		slog.Error(
			"Error occurred while calling xendit payment request api",
			"errs",
			errs,
			"resp",
			string(respByte),
		)
		return nil, errs[0]
	}

	if statusCode >= 400 && statusCode < 500 {
		var errMsg XenditErrMsg
		err := json.Unmarshal(respByte, &errMsg)
		if err != nil {
			return nil, err
		}
		slog.Error(
			"xendit payment request api returned 4xx status code",
			"code",
			statusCode,
			"resp",
			string(respByte),
		)

		if statusCode == 400 {
			return nil, status.Error(codes.InvalidArgument, errMsg.Message)
		}

		if statusCode == 404 {
			return nil, status.Error(codes.NotFound, errMsg.Message)
		}
	}

	var paymentRequestResponse XenditPaymentRequestResponse
	err := json.Unmarshal(respByte, &paymentRequestResponse)
	if err != nil {
		slog.Error(
			"Error occurred while unmarshalling xendit payment request api response",
			"err",
			err,
		)
		return nil, err
	}

	res := paymentRequestResponse.ToGetPaymentByIdGrpcRes()

	return res, nil
}

func (s *GrpcServer) Run() {
	ppb.RegisterPaymentServiceServer(s.Server, s)

	slog.Info("Starting gRPC server", "addr", s.ListenAddr)

	if err := s.Server.Serve(s.NetListener); err != nil {
		slog.Error("Error occurred while serving gRPC server", "err", err)
		panic(err)
	}
}

func getXenditCredentials() (string, string) {
	xenditApiKey := os.Getenv("XENDIT_API_KEY") + ":"
	xenditApiKeyBase64 := base64.StdEncoding.EncodeToString([]byte(xenditApiKey))
	xenditHost := os.Getenv("XENDIT_API_URL")
	paymentReqUrl := fmt.Sprintf("%s/v3/payment_requests/", xenditHost)

	return xenditApiKeyBase64, paymentReqUrl
}

func NewGrpcServer(listenAddr string, DB *sql.DB) *GrpcServer {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		slog.Error("Error occurred while creating listener", "err", err)
		panic(err)
	}

	return &GrpcServer{
		ListenAddr:  listenAddr,
		DB:          DB,
		Server:      grpc.NewServer(),
		NetListener: listener,
	}
}
