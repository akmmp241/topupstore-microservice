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
