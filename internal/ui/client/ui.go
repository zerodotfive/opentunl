package client

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/handler"
	"github.com/zerodotfive/opentunl/ui"
)

type StaticResponse struct {
	Status      int
	ContentType string `header:"Content-Type"`
	Body        []byte
}

func Register(hapi huma.API, container handler.Container) {
	huma.Register(hapi, huma.Operation{
		OperationID: "get-client-html-index",
		Method:      http.MethodGet,
		Path:        "/",
	}, func(ctx context.Context, request *struct{}) (*StaticResponse, error) {
		fileBytes, err := ui.GetEmbedHTMLBytes("client.html")
		if err != nil {
			return &StaticResponse{
				Status:      http.StatusInternalServerError,
				ContentType: "text/html",
				Body:        []byte("error opening file"),
			}, nil
		}
		resp := &StaticResponse{
			Status:      http.StatusOK,
			ContentType: "text/html",
			Body:        fileBytes,
		}
		return resp, nil
	})
}
