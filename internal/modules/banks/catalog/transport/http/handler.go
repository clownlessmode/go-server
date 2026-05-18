package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"project/internal/modules/banks/catalog/usecase/listbanks"
)

type Handler struct {
	listBanks *listbanks.UseCase
}

func NewHandler(listBanks *listbanks.UseCase) *Handler {
	return &Handler{listBanks: listBanks}
}

// ListBanks godoc
// @Summary List banks
// @Description Returns available banks.
// @Tags banks
// @Produce json
// @Success 200 {array} BankResponse
// @Failure 500 {object} BankErrorResponse
// @Router /banks [get]
func (h *Handler) ListBanks(c *gin.Context) {
	out, err := h.listBanks.Execute(c.Request.Context(), listbanks.Input{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	banks := make([]BankResponse, 0, len(out.Banks))
	for _, bank := range out.Banks {
		banks = append(banks, BankResponse{
			ID:        bank.ID,
			Code:      bank.Code,
			Name:      bank.Name,
			CreatedAt: bank.CreatedAt,
			UpdatedAt: bank.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, banks)
}
