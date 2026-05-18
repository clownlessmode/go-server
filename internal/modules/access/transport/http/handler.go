package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"project/internal/app/middleware"
	"project/internal/modules/access/domain"
	"project/internal/modules/access/usecase/grantaccess"
	"project/internal/modules/access/usecase/listmyaccesses"
	"project/internal/modules/access/usecase/revokeaccess"
	bankdomain "project/internal/modules/banks/catalog/domain"
	userdomain "project/internal/modules/user/domain"
)

type Handler struct {
	listMyAccesses *listmyaccesses.UseCase
	grantAccess    *grantaccess.UseCase
	revokeAccess   *revokeaccess.UseCase
}

func NewHandler(
	listMyAccesses *listmyaccesses.UseCase,
	grantAccess *grantaccess.UseCase,
	revokeAccess *revokeaccess.UseCase,
) *Handler {
	return &Handler{
		listMyAccesses: listMyAccesses,
		grantAccess:    grantAccess,
		revokeAccess:   revokeAccess,
	}
}

type GrantAccessRequest struct {
	UserID      int64     `json:"userId" binding:"required"`
	BankID      int64     `json:"bankId" binding:"required"`
	ExpiresAt   time.Time `json:"expiresAt" binding:"required"`
	GrantReason string    `json:"grantReason" binding:"required"`
}

type RevokeAccessRequest struct {
	UserID       int64  `json:"userId" binding:"required"`
	BankID       int64  `json:"bankId" binding:"required"`
	RevokeReason string `json:"revokeReason" binding:"required"`
}

// ListMyAccesses godoc
// @Summary List current user accesses
// @Description Returns current user's bank accesses. Admin grant reason is hidden from the user.
// @Tags accesses
// @Produce json
// @Security BearerAuth
// @Success 200 {array} AccessResponse
// @Failure 401 {object} AccessErrorResponse
// @Failure 500 {object} AccessErrorResponse
func (h *Handler) ListMyAccesses(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	out, err := h.listMyAccesses.Execute(c.Request.Context(), listmyaccesses.Input{
		UserID: userID,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	accesses := make([]AccessResponse, 0, len(out.Accesses))
	for _, access := range out.Accesses {
		accesses = append(accesses, AccessResponse{
			ID:           access.ID,
			UserID:       access.UserID,
			BankID:       access.BankID,
			BankCode:     access.BankCode,
			BankName:     access.BankName,
			GrantedAt:    access.GrantedAt,
			ExpiresAt:    access.ExpiresAt,
			RevokedAt:    access.RevokedAt,
			RevokeReason: access.RevokeReason,
			IsActive:     access.IsActive,
		})
	}

	c.JSON(http.StatusOK, accesses)
}

// GrantAccess godoc
// @Summary Grant bank access
// @Description Grants a bank access to a user. Admin must provide internal grant reason and expiration.
// @Tags accesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body GrantAccessRequest true "Grant access payload"
// @Success 201 {object} AccessResponse
// @Failure 400 {object} AccessErrorResponse
// @Failure 401 {object} AccessErrorResponse
// @Failure 403 {object} AccessErrorResponse
// @Failure 404 {object} AccessErrorResponse
// @Failure 409 {object} AccessErrorResponse
// @Failure 500 {object} AccessErrorResponse
func (h *Handler) GrantAccess(c *gin.Context) {
	var req GrantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.grantAccess.Execute(c.Request.Context(), grantaccess.Input{
		UserID:      req.UserID,
		BankID:      req.BankID,
		ExpiresAt:   req.ExpiresAt,
		GrantReason: req.GrantReason,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	grantReason := out.GrantReason
	c.JSON(http.StatusCreated, AccessResponse{
		ID:           out.ID,
		UserID:       out.UserID,
		BankID:       out.BankID,
		BankCode:     out.BankCode,
		BankName:     out.BankName,
		GrantedAt:    out.GrantedAt,
		ExpiresAt:    out.ExpiresAt,
		GrantReason:  &grantReason,
		RevokedAt:    out.RevokedAt,
		RevokeReason: out.RevokeReason,
		IsActive:     out.IsActive,
	})
}

// RevokeAccess godoc
// @Summary Revoke bank access
// @Description Revokes active user access to a bank and saves revoke reason.
// @Tags accesses
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body RevokeAccessRequest true "Revoke access payload"
// @Success 200 {object} AccessResponse
// @Failure 400 {object} AccessErrorResponse
// @Failure 401 {object} AccessErrorResponse
// @Failure 403 {object} AccessErrorResponse
// @Failure 404 {object} AccessErrorResponse
// @Failure 500 {object} AccessErrorResponse
func (h *Handler) RevokeAccess(c *gin.Context) {
	var req RevokeAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.revokeAccess.Execute(c.Request.Context(), revokeaccess.Input{
		UserID:       req.UserID,
		BankID:       req.BankID,
		RevokeReason: req.RevokeReason,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	grantReason := out.GrantReason
	c.JSON(http.StatusOK, AccessResponse{
		ID:           out.ID,
		UserID:       out.UserID,
		BankID:       out.BankID,
		BankCode:     out.BankCode,
		BankName:     out.BankName,
		GrantedAt:    out.GrantedAt,
		ExpiresAt:    out.ExpiresAt,
		GrantReason:  &grantReason,
		RevokedAt:    out.RevokedAt,
		RevokeReason: out.RevokeReason,
		IsActive:     out.IsActive,
	})
}

func currentUserID(c *gin.Context) (int64, bool) {
	value, ok := c.Get(middleware.CurrentUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing current user"})
		return 0, false
	}

	userID, ok := value.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid current user"})
		return 0, false
	}

	return userID, true
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidExpiration):
		c.JSON(http.StatusBadRequest, gin.H{"error": "expiresAt must be in the future"})
	case errors.Is(err, userdomain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, bankdomain.ErrBankNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "bank not found"})
	case errors.Is(err, domain.ErrAccessNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "access not found"})
	case errors.Is(err, domain.ErrAccessAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "active access already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
