package logger

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/zerodotfive/opentunl/internal/auth"
	"github.com/zerodotfive/opentunl/internal/handler"
)

func NewLoggerMiddleware(container handler.Container) func(ctx huma.Context, next func(huma.Context)) {
	jwt := auth.NewJWT(container.Config.JWTKey)

	return func(ctx huma.Context, next func(huma.Context)) {
		userName := "unauthenticated"
		if cookie, err := huma.ReadCookie(ctx, "session"); err == nil {
			if userInfo, err := jwt.Parse(cookie.Value); err == nil {
				userName = userInfo.Login
			}
		}

		next(ctx)

		container.Logger.Printf(
			"%s %s %s %d",
			userName,
			ctx.Method(),
			ctx.URL().Path,
			ctx.Status(),
		)
	}
}
