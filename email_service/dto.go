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
