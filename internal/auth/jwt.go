package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	key []byte
}

type UserInfo struct {
	Login  string   `json:"preferred_username"`
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
	jwt.RegisteredClaims
}

func NewJWT(key string) *JWT {
	return &JWT{
		key: []byte(key),
	}
}

func (j *JWT) Issue(u *UserInfo) (string, error) {
	c := UserInfo{
		Login: u.Login, Name: u.Name, Email: u.Email, Groups: u.Groups,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.Login,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(j.key)
}

func (j *JWT) Parse(raw string) (*UserInfo, error) {
	var c UserInfo
	_, err := jwt.ParseWithClaims(raw, &c,
		func(*jwt.Token) (any, error) { return j.key, nil },
		jwt.WithValidMethods([]string{"HS256"}), // защита от подмены алгоритма
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	if len(c.Login) == 0 {
		return nil, errors.New("invalid token")
	}

	return &c, nil
}
