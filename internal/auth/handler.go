package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fickysp/fynt-be-api/config"
	"github.com/fickysp/fynt-be-api/pkg/response"
	"github.com/fickysp/fynt-be-api/pkg/sentinel"
	"github.com/fickysp/fynt-be-api/pkg/utils"
	"github.com/fickysp/fynt-be-api/pkg/validation"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authUsecase AuthUsecase
	validator   *validation.StructValidator
	cfg         *config.Config
}

func NewAuthHandler(au AuthUsecase, val *validation.StructValidator, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authUsecase: au,
		validator:   val,
		cfg:         cfg,
	}
}

// @Summary Login
// @Description Melakukan login pengguna
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body DTOLoginRequest true "Payload login"
// @Success 200 {object} SwaggerLoginResponse "Login berhasil"
// @Failure 400 {object} response.SwaggerErrorResponse "Format JSON atau data tidak valid"
// @Failure 401 {object} response.SwaggerErrorResponse "Email atau password salah"
// @Failure 404 {object} response.SwaggerErrorResponse "User tidak ditemukan"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req DTOLoginRequest

	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind login request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format JSON tidak valid",
			Data:    nil,
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for login request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrValidate,
			Data:    nil,
		})
	}

	resp, err := h.authUsecase.Login(
		c.Context(),
		h.cfg.JWTSecret,
		h.cfg.RefreshTokenSecret,
		req,
	)

	if err != nil {
		slog.Error("Failed to login", "error", err.Error())
		if errors.Is(err, sentinel.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{
				Success: false,
				Message: "Email atau password salah",
				Data:    nil,
			})
		} else if errors.Is(err, sentinel.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.BaseResponse{
				Success: false,
				Message: "User tidak ditemukan",
				Data:    nil,
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
				Success: false,
				Message: sentinel.ErrInternalServerError.Error(),
				Data:    nil,
			})
		}
	}
	maxAgeRefresh := 24 * 3600
	if req.RememberMe {
		maxAgeRefresh = 7 * 24 * 3600
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		MaxAge:   maxAgeRefresh,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    resp.Role,
		HTTPOnly: true,
		Secure:   true,
		MaxAge:   maxAgeRefresh,
	})

	return c.Status(fiber.StatusOK).JSON(SwaggerLoginResponse{
		Success: true,
		Message: "Login berhasil",
		Data:    resp,
	})
}

// @Summary Register Seeker
// @Description Mendaftarkan akun seeker
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body DTOApplicantRegisterRequest true "Payload Registrasi Job Seeker"
// @Success 201 {object} response.BaseResponse "Registrasi berhasil"
// @Failure 400 {object} response.SwaggerErrorResponse "Body request atau data tidak valid"
// @Failure 409 {object} response.SwaggerErrorResponse "Alamat email sudah digunakan"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/auth/register/seeker [post]
func (h *AuthHandler) RegisterSeeker(c fiber.Ctx) error {
	var req DTOApplicantRegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format JSON tidak valid",
			Data:    nil,
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrValidate,
			Data:    nil,
		})
	}

	if err := h.authUsecase.RegisterApplicant(c.Context(), req); err != nil {
		slog.Error("Failed to register", "error", err.Error())
		if errors.Is(err, sentinel.ErrEmailIsUsed) {
			return c.Status(fiber.StatusConflict).JSON(response.BaseResponse{
				Success: false,
				Message: "Alamat email sudah digunakan",
				Data:    nil,
			})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Registrasi berhasil", Data: nil})
}

// @Summary Register Recruiter
// @Description Mendaftarkan akun recruiter
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body DTOCompanyRegisterRequest true "Payload Registrasi Recruiter / Perusahaan"
// @Success 201 {object} response.BaseResponse "Registrasi berhasil"
// @Failure 400 {object} response.SwaggerErrorResponse "Body request tidak valid"
// @Failure 409 {object} response.SwaggerErrorResponse "Alamat email sudah digunakan"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/auth/register/recruiter [post]
func (h *AuthHandler) RegisterRecruiter(c fiber.Ctx) error {
	var req DTOCompanyRegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		slog.Error("failed to bind register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Format JSON tidak valid", Data: nil})
	}

	if err := h.validator.Validate(&req); err != nil {
		slog.Error("Validation failed for register request", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrValidate, Data: nil})
	}

	if err := h.authUsecase.RegisterCompany(c.Context(), req); err != nil {
		slog.Error("Failed to register", "error", err.Error())
		if errors.Is(err, sentinel.ErrEmailIsUsed) {
			return c.Status(fiber.StatusConflict).JSON(response.BaseResponse{Success: false, Message: "Alamat email sudah digunakan", Data: nil})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(response.BaseResponse{Success: true, Message: "Registrasi berhasil", Data: nil})
}

// @Summary Logout
// @Description Mengeluarkan pengguna dari sesi aktif
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.BaseResponse "Logout berhasil"
// @Failure 401 {object} response.SwaggerErrorResponse "User tidak diketahui / tidak diautentikasi"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal memproses logout"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	accountID, ok := c.Locals("account_id").(string)
	slog.Info("account_id user", "account_id", accountID)
	if !ok || accountID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{Success: false, Message: "User tidak diketahui, silahkan login kembali", Data: nil})
	}
	if err := h.authUsecase.Logout(c.Context(), accountID); err != nil {
		slog.Error("Penyebab error", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
	})
	
	c.Cookie(&fiber.Cookie{
		Name:     "role",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   true,
	})

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{Success: true, Message: "Logout berhasil", Data: nil})
}

// @Summary Refresh Token
// @Description Membuat access token baru dari refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} SwaggerRefreshResponse "Berhasil mendapatkan access token"
// @Failure 400 {object} response.SwaggerErrorResponse "Refresh token tidak ditemukan"
// @Failure 401 {object} response.SwaggerErrorResponse "Unauthorized refresh token"
// @Failure 500 {object} response.SwaggerErrorResponse "Internal Server Error"
// @Router /api/v1/refresh-token [post]
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Refresh token tidak ditemukan", Data: nil})
	}

	result, err := h.authUsecase.RefreshToken(c.Context(), refreshToken, h.cfg.JWTSecret, h.cfg.RefreshTokenSecret)
	if err != nil {
		if errors.Is(err, sentinel.ErrUnauthorized) || errors.Is(err, sentinel.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{Success: false, Message: "User tidak diketahui, silahkan login kembali", Data: nil})
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{Success: false, Message: sentinel.ErrInternalServerError.Error(), Data: nil})
		}
	}

	return c.Status(fiber.StatusOK).JSON(SwaggerRefreshResponse{
		Success: true,
		Message: "Berhasil mendapatkan access token",
		Data:    result,
	})
}

// @Summary Login with Google
// @Description Mengarahkan pengguna ke halaman login Google
// @Tags Auth
// @Produce json
// @Param role query string true "User role (seeker/recruiter)"
// @Success 302 "Redirect to Google"
// @Failure 400 {object} response.SwaggerErrorResponse "Role user tidak valid"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal generate PKCE"
// @Router /api/v1/auth/google/login [get]
func (h *AuthHandler) LoginGoogle(c fiber.Ctx) error {
	role := c.Query("role")
	if role != "seeker" && role != "recruiter" {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Role user tidak valid", Data: nil})
	}
	state, _ := utils.RandomString()

	codeVerifier, codeChallenge, err := utils.GeneratePKCE()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: "Gagal generate PKCE",
			Data:    nil,
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		MaxAge:   600,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "pkce_verifier",
		Value:    codeVerifier,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		MaxAge:   600,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_role",
		Value:    role,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		MaxAge:   600,
	})

	url := GoogleOAuthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("code_challenge", codeChallenge), oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	return c.Redirect().To(url)
}

// @Summary Google OAuth Callback
// @Description Memproses callback autentikasi Google
// @Tags Auth
// @Produce json
// @Param state query string true "OAuth state"
// @Param code query string true "OAuth authorization code"
// @Success 302 "Redirect after authentication"
// @Failure 400 {object} response.SwaggerErrorResponse "Data state atau role tidak valid"
// @Failure 401 {object} response.SwaggerErrorResponse "User tidak diketahui"
// @Failure 500 {object} response.SwaggerErrorResponse "Gagal memproses profil Google"
// @Failure 502 {object} response.SwaggerErrorResponse "Layanan profil Google tidak tersedia"
// @Router /api/v1/auth/google/callback [get]
func (h *AuthHandler) CallbackGoogle(c fiber.Ctx) error {
	state := c.Query("state")
	code := c.Query("code")

	stateFromCookie := c.Cookies("oauth_state")
	if state != stateFromCookie {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Data state tidak valid", Data: nil})
	}

	codeVerifier := c.Cookies("pkce_verifier")
	token, err := GoogleOAuthConfig.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam(
			"code_verifier",
			codeVerifier,
		),
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{Success: false, Message: "Gagal mengubah kode", Data: nil})
	}

	role := c.Cookies("oauth_role")
	if role != string(RoleSeeker) && role != string(RoleRecruiter) {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Role OAuth tidak valid",
			Data:    nil,
		})
	}

	client := GoogleOAuthConfig.Client(c.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: "Gagal mengambil data profil Google",
			Data:    nil,
		})
	}
	defer resp.Body.Close()
	if resp.StatusCode < fiber.StatusOK || resp.StatusCode >= fiber.StatusMultipleChoices {
		return c.Status(fiber.StatusBadGateway).JSON(response.BaseResponse{
			Success: false,
			Message: "Layanan profil Google tidak tersedia",
			Data:    nil,
		})
	}

	var userInfo struct {
		Email         string `json:"email"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: "Gagal memproses informasi profil dari layanan autentikasi",
			Data:    nil,
		})
	}

	var internalJWT string

	if role == "recruiter" {
		internalJWT, err = h.authUsecase.OauthGoogleRecruiterLogin(c.Context(), h.cfg.RefreshTokenSecret, userInfo)
	} else {
		internalJWT, err = h.authUsecase.OauthGoogleSeekerLogin(c.Context(), h.cfg.RefreshTokenSecret, userInfo)
	}

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{Success: false, Message: "User tidak diketahui, silahkan login kembali", Data: nil})
	}

	exchangeCode, err := h.authUsecase.CreateExchangeCode(c.Context(), internalJWT)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrInternalServerError.Error(),
			Data:    nil,
		})
	}

	redirectURL := fmt.Sprintf("%s?code=%s", os.Getenv("FRONTEND_OAUTH_SUCCESS_URL"), exchangeCode)
	return c.Redirect().To(redirectURL)
}

// ExchangeToken handles exchanging one-time code for JWT token
// @Summary Exchange OAuth Code
// @Description Memvalidasi one-time code untuk mendapatkan access token JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body DTOExchangeCodeRequest true "Exchange Code Payload"
// @Success 200 {object} response.BaseResponse "Penukaran token berhasil"
// @Failure 400 {object} response.SwaggerErrorResponse "Format JSON tidak valid"
// @Failure 401 {object} response.SwaggerErrorResponse "Kode penukaran tidak valid atau sudah kadaluwarsa"
// @Router /api/v1/auth/exchange [post]
func (h *AuthHandler) ExchangeToken(c fiber.Ctx) error {
	var req DTOExchangeCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: "Format JSON tidak valid",
			Data:    nil,
		})
	}

	if err := h.validator.Validate(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.BaseResponse{
			Success: false,
			Message: sentinel.ErrValidate,
			Data:    nil,
		})
	}

	token, err := h.authUsecase.ExchangeCodeForToken(c.Context(), req.Code)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.BaseResponse{
			Success: false,
			Message: "Kode penukaran tidak valid atau sudah kadaluwarsa",
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.BaseResponse{
		Success: true,
		Message: "Penukaran token berhasil",
		Data: fiber.Map{
			"access_token": token,
		},
	})
}
