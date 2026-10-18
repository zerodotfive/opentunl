package nodes

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/node"
)

type ModelNodeDeleteResponse struct {
	Status int
	Body   *ModelNodeDeleteResponseBody
}

type ModelNodeDeleteResponseBody struct {
	Status bool `json:"status"`
}
type ModelNodeDeleteRequest struct {
	Name string `path:"name"`
}

var respDeleteError = &ModelNodeDeleteResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelNodeDeleteResponseBody{
		Status: false,
	},
}

func (h *Handler) delete(ctx context.Context, request *ModelNodeDeleteRequest) (*ModelNodeDeleteResponse, error) {
	provider := node.NewNodeProvider(ctx, h.postgres)

	err := provider.DeleteNode(request.Name)
	if err != nil {
		h.logger.Printf("error deleting node: %s", err.Error())
		return respDeleteError, errors.New("error deleting node")
	}

	return &ModelNodeDeleteResponse{
		Status: http.StatusOK,
		Body: &ModelNodeDeleteResponseBody{
			Status: true,
		},
	}, nil
}
