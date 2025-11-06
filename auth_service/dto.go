package main

import "time"

type RegisterRequest struct {
	Email                  string `json:"email" validate:"required,email"`
	Password               string `json:"password" validate:"required,min=8,max=255"`
	Name                   string `json:"name" validate:"required,min=2,max=255"`
	PhoneNumber            string `json:"phone_number" validate:"required,min=10,max=15"`
	PasswordConfirmation   string `json:"password_confirmation" validate:"required,eqfield=Password"`
	EmailVerificationToken string `json:"email_verification_token"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=255"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
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

type GetUserResponse struct {
	Message string       `json:"message"`
	Data    UserResponse `json:"data"`
	Errors  any          `json:"errors"`
}

type AuthEvent[T NewLoginMessage | NewRegistrationMessage | ForgotPasswordMessage] struct {
	EventTye string `json:"event_type"`
	Data     *T     `json:"data"`
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

type ForgotPasswordMessage struct {
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ResetUrl  string    `json:"reset_url"`
	ExpiresAt time.Time `json:"expired_at"`
}

type ResetPasswordRequest struct {
	ResetToken           string `json:"resetToken" validate:"required"`
	Password             string `json:"password" validate:"required,min=8,max=255"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,min=8,max=255"`
}

type UpdateUserRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Name        string `json:"name" validate:"required,min=2,max=255"`
	PhoneNumber string `json:"phone_number" validate:"required,min=10,max=15"`
	Password    string `json:"password" validate:"omitempty,min=8,max=255"`
}

type GetResponse struct {
	Id              int       `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	PhoneNumber     string    `json:"phone_number"`
	EmailVerifiedAt time.Time `json:"email_verified_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
