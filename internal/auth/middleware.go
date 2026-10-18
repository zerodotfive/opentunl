package auth

import (
	"net/http"
	"slices"

	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/handler"
	"github.com/zerodotfive/opentunl/ui"
)

func NewAdminAuthMiddleware(container handler.Container) func(ctx huma.Context, next func(huma.Context)) {
	jwt := NewJWT(container.Config.JWTKey)

	return func(ctx huma.Context, next func(huma.Context)) {
		cookie, err := huma.ReadCookie(ctx, "session")
		if err != nil {
			ctx.SetStatus(http.StatusUnauthorized)
			fileBytes, _ := ui.GetEmbedHTMLBytes("unauthorized.html")
			_, _ = ctx.BodyWriter().Write(fileBytes)
			return
		}

		userInfo, err := jwt.Parse(cookie.Value)

		if err != nil {
			ctx.SetStatus(http.StatusUnauthorized)
			fileBytes, _ := ui.GetEmbedHTMLBytes("unauthorized.html")
			_, _ = ctx.BodyWriter().Write(fileBytes)
			return
		}

		if !slices.Contains(userInfo.Groups, container.Config.AdminGroup) {
			ctx.SetStatus(http.StatusForbidden)
			fileBytes, _ := ui.GetEmbedHTMLBytes("forbidden.html")
			_, _ = ctx.BodyWriter().Write(fileBytes)
			return
		}

		next(ctx)
	}
}

func NewClientAuthMiddleware(container handler.Container) func(ctx huma.Context, next func(huma.Context)) {
	jwt := NewJWT(container.Config.JWTKey)

	return func(ctx huma.Context, next func(huma.Context)) {
		cookie, err := huma.ReadCookie(ctx, "session")
		if err != nil {
			ctx.SetStatus(http.StatusUnauthorized)
			fileBytes, _ := ui.GetEmbedHTMLBytes("unauthorized.html")
			_, _ = ctx.BodyWriter().Write(fileBytes)
			return
		}

		userInfo, err := jwt.Parse(cookie.Value)

		if err != nil {
			ctx.SetStatus(http.StatusUnauthorized)
			fileBytes, _ := ui.GetEmbedHTMLBytes("unauthorized.html")
			_, _ = ctx.BodyWriter().Write(fileBytes)
			return
		}

		next(huma.WithValue(ctx, "userInfo", userInfo))
	}
}
