package main

import "time"

type User struct {
	Id                     int
	Name                   string
	Email                  string
	Password               string
	PhoneNumber            string
	EmailVerificationToken string
	EmailVerifiedAt        time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
