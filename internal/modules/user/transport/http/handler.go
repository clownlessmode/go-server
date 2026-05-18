package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"project/internal/modules/user/domain"
	"project/internal/modules/user/usecase/createuser"
	"project/internal/modules/user/usecase/deleteuser"
	"project/internal/modules/user/usecase/getuser"
	"project/internal/modules/user/usecase/listusers"
	"project/internal/modules/user/usecase/updateuser"
)

type Handler struct {
	createUser *createuser.UseCase
	listUsers  *listusers.UseCase
	getUser    *getuser.UseCase
	updateUser *updateuser.UseCase
	deleteUser *deleteuser.UseCase
}

func NewHandler(
	createUser *createuser.UseCase,
	listUsers *listusers.UseCase,
	getUser *getuser.UseCase,
	updateUser *updateuser.UseCase,
	deleteUser *deleteuser.UseCase,
) *Handler {
	return &Handler{
		createUser: createUser,
		listUsers:  listUsers,
		getUser:    getUser,
		updateUser: updateUser,
		deleteUser: deleteUser,
	}
}

type CreateUserRequest struct {
	Login    string `json:"login" binding:"required"`
	Role     string `json:"role" binding:"required"`
	IsActive bool   `json:"isActive"`
}

type UpdateUserRequest struct {
	Login    *string `json:"login"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
	IsActive *bool   `json:"isActive"`
}

// CreateUser godoc
// @Summary Create user
// @Description Creates a user and generates a random 16-character password.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param input body CreateUserRequest true "User create payload"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	out, err := h.createUser.Execute(c.Request.Context(), createuser.Input{
		Login:    req.Login,
		Role:     domain.Role(req.Role),
		IsActive: req.IsActive,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, userResponse(userResponseInput{
		ID:        out.ID,
		Login:     out.Login,
		Password:  out.Password,
		Role:      out.Role,
		IsActive:  out.IsActive,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}))
}

// ListUsers godoc
// @Summary List users
// @Description Returns all users with visible passwords.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func (h *Handler) ListUsers(c *gin.Context) {
	out, err := h.listUsers.Execute(c.Request.Context(), listusers.Input{})
	if err != nil {
		handleError(c, err)
		return
	}

	users := make([]UserResponse, 0, len(out.Users))
	for _, user := range out.Users {
		users = append(users, userResponse(userResponseInput{
			ID:        user.ID,
			Login:     user.Login,
			Password:  user.Password,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}))
	}

	c.JSON(http.StatusOK, users)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Returns one user with visible password.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func (h *Handler) GetUser(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	out, err := h.getUser.Execute(c.Request.Context(), getuser.Input{
		ID: id,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, userResponse(userResponseInput{
		ID:        out.ID,
		Login:     out.Login,
		Password:  out.Password,
		Role:      out.Role,
		IsActive:  out.IsActive,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}))
}

// UpdateUser godoc
// @Summary Update user
// @Description Updates login, password, role and active flag. ID and timestamps cannot be changed by request body.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param input body UpdateUserRequest true "User update payload"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func (h *Handler) UpdateUser(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	input := updateuser.Input{
		ID:       id,
		Login:    req.Login,
		Password: req.Password,
		IsActive: req.IsActive,
	}
	if req.Role != nil {
		role := domain.Role(*req.Role)
		input.Role = &role
	}

	out, err := h.updateUser.Execute(c.Request.Context(), input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, userResponse(userResponseInput{
		ID:        out.ID,
		Login:     out.Login,
		Password:  out.Password,
		Role:      out.Role,
		IsActive:  out.IsActive,
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Deletes a user by ID.
// @Tags users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
func (h *Handler) DeleteUser(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	_, err := h.deleteUser.Execute(c.Request.Context(), deleteuser.Input{
		ID: id,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func parseIDParam(c *gin.Context) (int64, bool) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return 0, false
	}

	return id, true
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
	case errors.Is(err, domain.ErrInvalidRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user role"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

type userResponseInput struct {
	ID        int64
	Login     string
	Password  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func userResponse(input userResponseInput) UserResponse {
	return UserResponse{
		ID:        input.ID,
		Login:     input.Login,
		Password:  input.Password,
		Role:      input.Role,
		IsActive:  input.IsActive,
		CreatedAt: input.CreatedAt,
		UpdatedAt: input.UpdatedAt,
	}
}
