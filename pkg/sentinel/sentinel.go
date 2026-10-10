package sentinel

import "errors"

var (
	ErrInternalServerError = errors.New("Internal Server Error")
	ErrInvalidCredentials  = errors.New("Email or Password Incorrect")
	ErrEmailIsUsed         = errors.New("Email has already been used")
	ErrUserNotFound        = errors.New("User not found")
)
