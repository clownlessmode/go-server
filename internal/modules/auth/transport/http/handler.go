package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	authdomain "project/internal/modules/auth/domain"
	"project/internal/modules/auth/usecase/login"
	"project/internal/modules/auth/usecase/logout"
	"project/internal/modules/auth/usecase/refresh"
)

type Handler struct {
	login   *login.UseCase
	refresh *refresh.UseCase
	logout  *logout.UseCase
}

func NewHandler(login *login.UseCase, refresh *refresh.UseCase, logout *logout.UseCase) *Handler {
	return &Handler{
		login:   login,
		refresh: refresh,
		logout:  logout,
	}
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// Login godoc
// @Summary Login
// @Description Authenticates a user and returns user, access token and refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LoginRequest true "Login payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.login.Execute(c.Request.Context(), login.Input{
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse(authResponseInput{
		User: authUserResponseInput{
			ID:        out.User.ID,
			Login:     out.User.Login,
			Password:  out.User.Password,
			Role:      out.User.Role,
			IsActive:  out.User.IsActive,
			CreatedAt: out.User.CreatedAt,
			UpdatedAt: out.User.UpdatedAt,
		},
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}))
}

// Refresh godoc
// @Summary Refresh tokens
// @Description Rotates refresh token and returns user, new access token and new refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body RefreshRequest true "Refresh payload"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.refresh.Execute(c.Request.Context(), refresh.Input{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse(authResponseInput{
		User: authUserResponseInput{
			ID:        out.User.ID,
			Login:     out.User.Login,
			Password:  out.User.Password,
			Role:      out.User.Role,
			IsActive:  out.User.IsActive,
			CreatedAt: out.User.CreatedAt,
			UpdatedAt: out.User.UpdatedAt,
		},
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}))
}

// Logout godoc
// @Summary Logout
// @Description Revokes the provided refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Param input body LogoutRequest true "Logout payload"
// @Success 204
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
func (h *Handler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if _, err := h.logout.Execute(c.Request.Context(), logout.Input{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
	case errors.Is(err, authdomain.ErrInactiveUser):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user is inactive"})
	case errors.Is(err, authdomain.ErrInvalidToken), errors.Is(err, authdomain.ErrRefreshRevoked):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

type authResponseInput struct {
	User         authUserResponseInput
	AccessToken  string
	RefreshToken string
}

type authUserResponseInput struct {
	ID        int64
	Login     string
	Password  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func authResponse(input authResponseInput) AuthResponse {
	return AuthResponse{
		User: AuthUserResponse{
			ID:        input.User.ID,
			Login:     input.User.Login,
			Password:  input.User.Password,
			Role:      input.User.Role,
			IsActive:  input.User.IsActive,
			CreatedAt: input.User.CreatedAt,
			UpdatedAt: input.User.UpdatedAt,
		},
		AccessToken:  input.AccessToken,
		RefreshToken: input.RefreshToken,
	}
}
