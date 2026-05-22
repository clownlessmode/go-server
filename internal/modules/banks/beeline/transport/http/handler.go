package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"project/internal/app/logger"
	"project/internal/modules/banks/beeline/usecase/getconfig"
)

var handlerLog = logger.New("beeline-http")

type Handler struct {
	getConfig *getconfig.UseCase
}

func NewHandler(getConfig *getconfig.UseCase) *Handler {
	return &Handler{getConfig: getConfig}
}

// GetConfig godoc
// @Summary Get Beeline config
// @Description Returns Beeline config.
// @Tags beeline config
// @Produce json
// @Success 200 {object} ConfigResponse
// @Failure 500 {object} BeelineErrorResponse
// @Router /banks/beeline/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	out, err := h.getConfig.Execute(c.Request.Context(), getconfig.Input{})
	if err != nil {
		handlerLog.Errorf("get config failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, ConfigResponse{
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	})
}
