package auth

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleSeeker    Role = "seeker"
	RoleRecruiter Role = "recruiter"
)

type Account struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	Email         string       `db:"email" json:"email"`
	Password      *string      `db:"password" json:"-"`
	PhoneNumber   *string      `db:"phone_number" json:"phone_number"`
	Role          Role         `db:"role" json:"role"`
	AuthProvider  *string      `db:"auth_provider" json:"auth_provider,omitempty"`
	RefreshToken  *string      `db:"refresh_token" json:"-"`
	EmailVerified bool         `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt     sql.NullTime `db:"deleted_at" json:"delete_at,omitempty"`
}

// DTO
// Request DTO
type DTOApplicantRegisterRequest struct {
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	PhoneNumber    string `json:"phone_number" validate:"required"`
	Password       string `json:"password" validate:"required,min=8"`
	VerifyPassword string `json:"verify_password" validate:"required,eqfield=Password"`
	CallbackUrl    string `json:"callback_url"`
}

type DTOCompanyRegisterRequest struct {
	CompanyName    string `json:"company_name" validate:"required"`
	Sector         string `json:"sector" validate:"required"`
	Domicile       string `json:"domicile" validate:"required"`
	WebURL         string `json:"web_url"`
	Address        string `json:"address" validate:"required"`
	OwnerName      string `json:"owner_name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	PhoneNumber    string `json:"phone_number" validate:"required"`
	Password       string `json:"password" validate:"required,min=8" `
	VerifyPassword string `json:"verify_password" validate:"required,eqfield=Password"`
	CallbackUrl    string `json:"callback_url"`
}

type DTOLoginRequest struct {
	Email      string `json:"email" validate:"required,email" `
	Password   string `json:"password" validate:"required,min=8" `
	RememberMe bool   `json:"remember_me"`
}

type DTOSSORequest struct {
	Provider string `validate:"required" json:"provider"`
	Token    string `validate:"required" json:"token"`
}
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

type ResendVerificationEmailRequest struct {
	Email       string `json:"email" validate:"required,email"`
	CallbackURL string `json:"callback_url"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// Response DTO
type DataLogin struct {
	Profile any `json:"profile"`
}
type DTOLoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"-"`
	Data         DataLogin `json:"data"`
}

type DTORefreshResponse struct {
	AccessToken string `json:"access_token"`
}
