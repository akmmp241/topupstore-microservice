package main

import "time"

type User struct {
	Id                     int       `json:"id"`
	Name                   string    `json:"name"`
	Email                  string    `json:"email"`
	Password               string    `json:"password"`
	PhoneNumber            string    `json:"phone_number"`
	EmailVerificationToken string    `json:"email_verification_token"`
	EmailVerifiedAt        time.Time `json:"email_verified_at"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
