package login

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zerodotfive/opentunl/internal/auth"
	"golang.org/x/oauth2"
)

type ModelSessionInfoGETResponse struct {
	Status int
	Body   *ModelSessionInfoGETResponseBody
}

type ModelSessionInfoGETResponseBody struct {
	Status   bool           `json:"status"`
	UserInfo *auth.UserInfo `json:"user_info"`
}

type ModelSessionInfoGETRequest struct {
	SessionCookie http.Cookie `cookie:"session" required:"true"`
}

type ModelAuthResponse struct {
	Status   int
	Location string        `header:"Location"`
	Cookies  []http.Cookie `header:"Set-Cookie"`
}

type ModelLoginGETRequest struct {
	Return string `query:"return"`
}
type ModelLoginCallbackGETRequest struct {
	ReturnCookie string      `cookie:"return" required:"true"`
	StateCookie  http.Cookie `cookie:"state" required:"true"`
	Code         string      `query:"code" required:"true"`
	State        string      `query:"state" required:"true"`
}

func (h *Handler) login(ctx context.Context, request *ModelLoginGETRequest) (*ModelAuthResponse, error) {
	_, oauth, err := auth.GetOAuth(ctx, h.config)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 16)
	_, err = rand.Read(b)
	if err != nil {
		return &ModelAuthResponse{
			Status: http.StatusInternalServerError,
		}, nil
	}

	state := base64.RawURLEncoding.EncodeToString(b)

	return &ModelAuthResponse{
		Status:   http.StatusFound,
		Location: oauth.AuthCodeURL(state, oauth2.SetAuthURLParam("return", request.Return)),
		Cookies: []http.Cookie{
			{
				Name:     "state",
				Value:    state,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   300,
			},
			{
				Name:     "return",
				Value:    request.Return,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   300,
			},
			{
				Name:     "session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			},
		},
	}, nil
}

func (h *Handler) logout(ctx context.Context, request *ModelLoginGETRequest) (*ModelAuthResponse, error) {
	return &ModelAuthResponse{
		Status:   http.StatusFound,
		Location: request.Return,
		Cookies: []http.Cookie{
			{
				Name:     "state",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			},
			{
				Name:     "session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			},
		},
	}, nil
}

func (h *Handler) callback(ctx context.Context, request *ModelLoginCallbackGETRequest) (*ModelAuthResponse, error) {
	provider, oauth, err := auth.GetOAuth(ctx, h.config)
	if err != nil {
		return nil, errors.New("oauth failed")
	}

	if request.State == "" || request.State != request.StateCookie.Value {
		return &ModelAuthResponse{Status: http.StatusBadRequest}, nil
	}

	token, err := oauth.Exchange(ctx, request.Code)
	if err != nil {
		return &ModelAuthResponse{Status: http.StatusBadRequest}, nil
	}

	resp, err := oauth.Client(ctx, token).Get(provider.UserInfoEndpoint())
	if err != nil {
		return &ModelAuthResponse{Status: http.StatusBadRequest}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	var userInfo auth.UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return &ModelAuthResponse{Status: http.StatusBadRequest}, nil
	}

	jwt := auth.NewJWT(h.config.JWTKey)
	session, err := jwt.Issue(&userInfo)

	return &ModelAuthResponse{
		Status:   http.StatusFound,
		Location: request.ReturnCookie,
		Cookies: []http.Cookie{
			{
				Name:     "state",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			},
			{
				Name:     "return",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			},
			{
				Name:     "session",
				Value:    session,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   3600,
			},
		},
	}, nil
}

func (h *Handler) info(ctx context.Context, request *ModelSessionInfoGETRequest) (*ModelSessionInfoGETResponse, error) {
	jwt := auth.NewJWT(h.config.JWTKey)
	userInfo, err := jwt.Parse(request.SessionCookie.Value)
	if err != nil {
		return &ModelSessionInfoGETResponse{Status: http.StatusUnauthorized}, nil
	}

	return &ModelSessionInfoGETResponse{
		Status: http.StatusOK,
		Body: &ModelSessionInfoGETResponseBody{
			Status:   true,
			UserInfo: userInfo,
		},
	}, nil
}
