package health

import (
	"context"
	"errors"
	"net/http"
)

type ModelHealthGETResponse struct {
	Status int
	Body   *ModelHealthGETResponseBody
}

type ModelHealthGETResponseBody struct {
	Status bool `json:"status"`
}

type ModelHealthGETRequest struct{}

func (h *Handler) get(ctx context.Context, request *ModelHealthGETRequest) (*ModelHealthGETResponse, error) {
	if err := h.postgres.Ping(ctx); err != nil {
		h.logger.Printf(err.Error())
		return &ModelHealthGETResponse{
			Status: http.StatusInternalServerError,
			Body: &ModelHealthGETResponseBody{
				Status: false,
			},
		}, errors.New("failed to dial database")
	}

	return &ModelHealthGETResponse{
		Status: http.StatusOK,
		Body: &ModelHealthGETResponseBody{
			Status: true,
		},
	}, nil
}
