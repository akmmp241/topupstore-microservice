package main

import (
	"time"

	ppb "github.com/akmmp241/topupstore-microservice/payment-proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CreatePaymentRequest struct {
	ReferenceId       string  `json:"reference_id"        validate:"required"`
	ChannelCode       string  `json:"channel_code"        validate:"required"`
	Amount            float64 `json:"amount"              validate:"required,min=1"`
	BuyerEmail        string  `json:"buyer_email"         validate:"required,email"`
	BuyerMobileNumber string  `json:"buyer_mobile_number" validate:"omitempty"`
}

type CreatePaymentResponse struct {
	XenditPaymentId string `json:"xendit_payment_id" validate:"required"`
	Status          string `json:"status"            validate:"required"`
	FailureCode     string `json:"failure_code"`
}

type XenditPaymentRequestResponse struct {
	ReferenceId       string            `json:"reference_id"       validate:"required"`
	PaymentRequestId  string            `json:"payment_request_id" validate:"required"`
	PaymentTokenId    string            `json:"payment_token_id"   validate:"required"`
	LatestPaymentId   string            `json:"latest_payment_id"  validate:"required"`
	Type              string            `json:"type"               validate:"required"`
	RequestAmount     int               `json:"request_amount"     validate:"required"`
	CaptureMethod     string            `json:"capture_method"     validate:"required"`
	ChannelCode       string            `json:"channel_code"       validate:"required"`
	ChannelProperties ChannelProperties `json:"channel_properties" validate:"required"`
	Country           string            `json:"country"            validate:"required"`
	Currency          string            `json:"currency"           validate:"required"`
	Actions           []Action          `json:"actions"            validate:"required"`
	Status            string            `json:"status"             validate:"required"`
	FailureCode       string            `json:"failure_code"`
	Created           time.Time         `json:"created"            validate:"required"`
	Updated           time.Time         `json:"updated"            validate:"required"`
}

func (x *XenditPaymentRequestResponse) ToGetPaymentByIdGrpcRes() *ppb.GetPaymentByIdRes {
	var actions []*ppb.Action
	for _, action := range x.Actions {
		actions = append(actions, &ppb.Action{
			Type:        action.Type,
			Descriptor_: action.Descriptor,
			Value:       action.Value,
		})
	}

	return &ppb.GetPaymentByIdRes{
		PaymentRequestId: x.PaymentRequestId,
		RequestAmount:    int32(x.RequestAmount),
		ChannelCode:      x.ChannelCode,
		ChannelProperties: &ppb.ChannelProperties{
			DisplayName:      x.ChannelProperties.DisplayName,
			ExpiresAt:        timestamppb.New(x.ChannelProperties.ExpiresAt),
			SuccessReturnUrl: x.ChannelProperties.SuccessReturnUrl,
			FailureReturnUrl: x.ChannelProperties.FailureReturnUrl,
			CancelReturnUrl:  x.ChannelProperties.CancelReturnUrl,
		},
		Actions:     actions,
		Status:      x.Status,
		FailureCode: x.FailureCode,
		Created:     timestamppb.New(x.Created),
		Updated:     timestamppb.New(x.Updated),
	}
}

type ChannelProperties struct {
	DisplayName      string    `json:"display_name,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	SuccessReturnUrl string    `json:"success_return_url,omitempty"`
	FailureReturnUrl string    `json:"failure_return_url,omitempty"`
	CancelReturnUrl  string    `json:"cancel_return_url,omitempty"`
}

type XenditRequestBody struct {
	Currency          string            `json:"currency"           validate:"required"`
	Type              string            `json:"type"`
	RequestAmount     int               `json:"request_amount"     validate:"required,min=1"`
	Country           string            `json:"country"`
	CaptureMethod     string            `json:"capture_method"`
	ReferenceId       string            `json:"reference_id"       validate:"required"`
	ChannelCode       string            `json:"channel_code"`
	ChannelProperties ChannelProperties `json:"channel_properties"`
}

type Action struct {
	Type       string `json:"type"       validate:"required"`
	Descriptor string `json:"descriptor" validate:"required"`
	Value      string `json:"value"      validate:"required"`
}

type GetPaymentByIdResponse struct {
	PaymentRequestId  string            `json:"payment_request_id" validate:"required"`
	RequestAmount     int               `json:"request_amount"     validate:"required"`
	ChannelCode       string            `json:"channel_code"       validate:"required"`
	ChannelProperties ChannelProperties `json:"channel_properties" validate:"required"`
	Actions           []Action          `json:"actions"            validate:"required"`
	Status            string            `json:"status"             validate:"required"`
	FailureCode       string            `json:"failure_code"`
	Created           time.Time         `json:"created"            validate:"required"`
	Updated           time.Time         `json:"updated"            validate:"required"`
}

type XenditErrMsg struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}
