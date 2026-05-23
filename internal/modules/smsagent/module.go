package smsagent

import (
	"github.com/gin-gonic/gin"

	agenthttp "project/internal/modules/smsagent/transport/http"
	"project/internal/modules/smsagent/usecase/ackmessage"
	"project/internal/modules/smsagent/usecase/listpending"
	agentdomain "project/internal/modules/smsagent/domain"
)

type Module struct {
	handler *agenthttp.Handler
	apiKey  string
}

func NewModule(repo agentdomain.Repository, apiKey string) *Module {
	return &Module{
		handler: agenthttp.NewHandler(
			listpending.New(repo),
			ackmessage.New(repo),
		),
		apiKey: apiKey,
	}
}

func (m *Module) RegisterRoutes(router *gin.Engine) {
	agenthttp.RegisterRoutes(router, m.handler, m.apiKey)
}
