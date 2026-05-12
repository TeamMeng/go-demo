package web

import (
	"time"

	"github.com/TeamMeng/go-demo/webook/internal/web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtHandler struct {
	atKey []byte
	rtKey []byte
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid  int64
	Ssid string
}

func newJwtHandler() jwtHandler {
	return jwtHandler{
		atKey: []byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISP"),
		rtKey: []byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISL"),
	}
}

func (h jwtHandler) setJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	claim := middleware.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
		Ssid:      ssid,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	tokenStr, err := token.SignedString(h.atKey)
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h jwtHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claim := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	tokenStr, err := token.SignedString(h.rtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h jwtHandler) setLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	if err := h.setJWTToken(ctx, uid, ssid); err != nil {
		return err
	}

	err := h.setRefreshToken(ctx, uid, ssid)
	return err
}
