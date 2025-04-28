package main

import "time"

type SendMail struct {
	To      string
	Subject string
	Body    string
}

type NewLoginMessage struct {
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	LoginTime time.Time `json:"login_time"`
	IpAddress string    `json:"ip_address"`
	Device    string    `json:"device"`
}

type NewRegistrationMessage struct {
	Name            string `json:"name"`
	VerificationUrl string `json:"verification_url"`
	Email           string `json:"email"`
}
