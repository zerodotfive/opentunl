package nodes

import (
	"context"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/node"
)

type ModelNodeListResponse struct {
	Status int
	Body   *ModelNodeListResponseBody
}

type ModelNodeListResponseBody struct {
	Status bool        `json:"status"`
	Nodes  []node.Node `json:"nodes"`
}
type ModelNodeListRequest struct{}

var respListError = &ModelNodeListResponse{
	Status: http.StatusInternalServerError,
	Body: &ModelNodeListResponseBody{
		Status: false,
	},
}

func (h *Handler) list(ctx context.Context, request *ModelNodeListRequest) (*ModelNodeListResponse, error) {
	provider := node.NewNodeProvider(ctx, h.postgres)

	nodes, err := provider.ListNodes()
	if err != nil {
		h.logger.Printf("error listing nodes: %s", err.Error())
		return respListError, errors.New("error listing nodes")
	}

	return &ModelNodeListResponse{
		Status: http.StatusOK,
		Body: &ModelNodeListResponseBody{
			Status: true,
			Nodes:  nodes,
		},
	}, nil
}
