package main

import (
	"net/smtp"
	"os"
)

type Mailer struct {
	Auth smtp.Auth
}

func NewMailer() *Mailer {
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")

	return &Mailer{Auth: smtp.PlainAuth("", username, password, host)}
}

func (m *Mailer) SendMail(data *SendMail) error {
	from := os.Getenv("SMTP_FROM")
	port := os.Getenv("SMTP_PORT")
	host := os.Getenv("SMTP_HOST")

	msg := "From: " + from + "\n" +
		"To: " + data.To + "\n" +
		"Subject: " + data.Subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		data.Body

	err := smtp.SendMail(host+":"+port, m.Auth, from, []string{data.To}, []byte(msg))
	return err
}
