package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"html/template"
	"log/slog"
	"os"
)

//go:embed templates/new-login.html
var NewLoginEmail string

//go:embed templates/user-registration.html
var NewRegistrationEmail string

//go:embed templates/forget-password.html
var ForgetPasswordEmail string

//go:embed templates/new-order.html
var NewOrderEmail string

type EmailService struct {
	Mailer *Mailer
}

func NewEmailService(mailer *Mailer) *EmailService {
	return &EmailService{Mailer: mailer}
}

func (e *EmailService) HandleUserRegistration(msg *kafka.Message) error {

	newUserMsg := NewRegistrationMessage{}
	if err := json.Unmarshal(msg.Value, &newUserMsg); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	tmpl, err := template.New("user-registration").Parse(NewRegistrationEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, newUserMsg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = newUserMsg.Email
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

func (e *EmailService) HandleUserLogin(msg *kafka.Message) error {

	newLoginMsg := NewLoginMessage{}
	if err := json.Unmarshal(msg.Value, &newLoginMsg); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	tmpl, err := template.New("new-login").Parse(NewLoginEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, newLoginMsg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = newLoginMsg.Email
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

func (e *EmailService) HandleForgotPassword(msg *kafka.Message) error {
	forgotPasswordMsg := ForgotPasswordMessage{}
	if err := json.Unmarshal(msg.Value, &forgotPasswordMsg); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	tmpl, err := template.New("forgot-password").Parse(ForgetPasswordEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, forgotPasswordMsg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = forgotPasswordMsg.Email
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

func (e *EmailService) HandleNewOrder(msg *kafka.Message) error {
	newOrderMsg := NewOrderMsg{}
	if err := json.Unmarshal(msg.Value, &newOrderMsg); err != nil {
		slog.Error("Error unmarshalling message", "error", err)
		return err
	}

	tmpl, err := template.New("new-order").Parse(NewOrderEmail)
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, newOrderMsg); err != nil {
		slog.Error("Error creating buffer", "error", err)
		return err
	}

	to := os.Getenv("SMTP_FROM")
	if os.Getenv("APP_ENV") == "production" {
		to = newOrderMsg.BuyerEmail
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
