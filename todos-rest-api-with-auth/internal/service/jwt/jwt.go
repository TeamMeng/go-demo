package jwt

import (
	"time"
	"todo_api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	AtKey []byte `json:"at_key"`
	RtKey []byte `json:"rt_key"`
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       string
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid string
}

func NewJWTHandler(cfg *config.Config) *JWTHandler {
	return &JWTHandler{
		AtKey: []byte(cfg.JWTSecret),
		RtKey: []byte(cfg.JWTSecret),
	}
}

func (h JWTHandler) SetJwtToken(ctx *gin.Context, uid string) error {
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.AtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h JWTHandler) SetRefreshToken(ctx *gin.Context, uid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(h.RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
