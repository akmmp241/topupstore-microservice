package main

type RegisterRequest struct {
	Email                  string `json:"email" validate:"required,email"`
	Password               string `json:"password" validate:"required,min=8,max=255"`
	Name                   string `json:"name" validate:"required,min=2,max=255"`
	PhoneNumber            string `json:"phone_number" validate:"required,min=10,max=15"`
	PasswordConfirmation   string `json:"password_confirmation" validate:"required,eqfield=Password"`
	EmailVerificationToken string `json:"email_verification_token" validate:"required"`
}
