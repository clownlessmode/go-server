package http

import "time"

type ConfigResponse struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BeelineErrorResponse struct {
	Error string `json:"error"`
}
