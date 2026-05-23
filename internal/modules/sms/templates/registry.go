package templates

import (
	"project/internal/modules/sms/domain"
)

type Renderer func(data any) (domain.Message, error)

type Registry struct {
	renderers map[domain.BankCode]Renderer
}

func NewRegistry() *Registry {
	registry := &Registry{
		renderers: make(map[domain.BankCode]Renderer),
	}
	registry.Register(domain.BankBeeline, RenderBeelinePayment)

	return registry
}

func (r *Registry) Register(bank domain.BankCode, renderer Renderer) {
	r.renderers[bank] = renderer
}

func (r *Registry) Render(bank domain.BankCode, data any) (domain.Message, error) {
	renderer, ok := r.renderers[bank]
	if !ok {
		return domain.Message{}, domain.ErrUnsupportedBank
	}

	return renderer(data)
}
