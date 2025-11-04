package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"os"

	"github.com/segmentio/kafka-go"
)

//go:embed templates/new-login.html
var NewLoginEmail string

//go:embed templates/user-registration.html
var NewRegistrationEmail string

//go:embed templates/forget-password.html
var ForgetPasswordEmail string

//go:embed templates/new-order.html
var NewOrderEmail string

//go:embed templates/success-order.html
var SuccessOrderEmail string

//go:embed templates/failed-order.html
var FailedOrderEmail string

const (
	UserRegistration = "user-registration"
	UserLogin        = "user-login"
	ForgotPassword   = "forgot-password"
	NewOrder         = "new-order"
	SuccessOrder     = "order-succeeded"
	FailedOrder      = "order-failed"
)

type EmailService struct {
	Mailer *Mailer
}

func NewEmailService(mailer *Mailer) *EmailService {
	return &EmailService{Mailer: mailer}
}

func (e *EmailService) HandleAuth(msg *kafka.Message) error {
	var base BaseEvent

	if err := json.Unmarshal(msg.Value, &base); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	switch base.EventType {
	case UserRegistration:
		var data *AuthEvent[NewRegistrationMessage]
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return err
		}
		return e.handleUserRegistration(data.Data)
	case UserLogin:
		var data *AuthEvent[NewLoginMessage]
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return err
		}
		return e.handleUserLogin(data.Data)
	case ForgotPassword:
		var data *AuthEvent[ForgotPasswordMessage]
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return err
		}
		return e.handleForgotPassword(data.Data)
	default:
		slog.Warn("Unknown event type", "event-type", base.EventType)
		return nil
	}
}

func (e *EmailService) HandleOrder(msg *kafka.Message) error {
	var base BaseEvent

	if err := json.Unmarshal(msg.Value, &base); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	switch base.EventType {
	case NewOrder:
		var data *OrderEvent
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return err
		}
		return e.handleNewOrder(data.Data)
	case SuccessOrder:
		var data *OrderEvent
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return err
		}
		return e.handleSuccessOrder(data.Data)
	case FailedOrder:
		var data *OrderEvent
		if err := json.Unmarshal(msg.Value, &data); err != nil {
			slog.Error("Error unmarshalling message", "error", err)
			return err
		}
		return e.handleFailedOrder(data.Data)
	default:
		slog.Warn("Unknown event type", "event-type", base.EventType)
		return nil
	}
}

func (e *EmailService) handleUserRegistration(msg *NewRegistrationMessage) error {
	tmpl, err := template.New("user-registration").Parse(NewRegistrationEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = msg.Email
	}

	emailData := &SendMail{
		To:      to,
		Subject: "User Registration Confirmation",
		Body:    body.String(),
	}

	if err := e.Mailer.SendMail(emailData); err != nil {
		slog.Error("Error sending mail", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", emailData.Subject)

	return nil
}

func (e *EmailService) handleUserLogin(msg *NewLoginMessage) error {
	tmpl, err := template.New("new-login").Parse(NewLoginEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = msg.Email
	}

	emailData := &SendMail{
		To:      to,
		Subject: "New Login Alert",
		Body:    body.String(),
	}

	if err := e.Mailer.SendMail(emailData); err != nil {
		slog.Error("Error sending mail", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", emailData.Subject)

	return nil
}

func (e *EmailService) handleForgotPassword(msg *ForgotPasswordMessage) error {
	tmpl, err := template.New("forgot-password").Parse(ForgetPasswordEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = msg.Email
	}

	emailData := &SendMail{
		To:      to,
		Subject: "Forgot Password",
		Body:    body.String(),
	}

	if err := e.Mailer.SendMail(emailData); err != nil {
		slog.Error("Error sending mail", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", emailData.Subject)

	return nil
}

func (e *EmailService) handleNewOrder(msg *OrderMsg) error {
	tmpl, err := template.New("new-order").Parse(NewOrderEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = msg.BuyerEmail
	}

	emailData := &SendMail{
		To:      to,
		Subject: "New Order Confirmation",
		Body:    body.String(),
	}

	if err := e.Mailer.SendMail(emailData); err != nil {
		slog.Error("Error sending mail", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", emailData.Subject)

	return nil
}

func (e *EmailService) handleSuccessOrder(msg *OrderMsg) error {
	tmpl, err := template.New("order-succeeded").Parse(SuccessOrderEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = msg.BuyerEmail
	}

	emailData := &SendMail{
		To:      to,
		Subject: "Order Payment Succeeded",
		Body:    body.String(),
	}

	if err := e.Mailer.SendMail(emailData); err != nil {
		slog.Error("Error sending mail", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", emailData.Subject)

	return nil
}

func (e *EmailService) handleFailedOrder(msg *OrderMsg) error {
	tmpl, err := template.New("order-failed").Parse(FailedOrderEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, msg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = msg.BuyerEmail
	}

	emailData := &SendMail{
		To:      to,
		Subject: "Order Payment Failed",
		Body:    body.String(),
	}

	if err := e.Mailer.SendMail(emailData); err != nil {
		slog.Error("Error sending mail", "error", err)
		return err
	}

	slog.Info("Email sent successfully", "to", to, "subject", emailData.Subject)

	return nil
}
