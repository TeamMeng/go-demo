package web

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type jwtHandler struct {
	atKey []byte
	rtKey []byte
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid int64
}

func newJwtHandler() jwtHandler {
	return jwtHandler{
		atKey: []byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISP"),
		rtKey: []byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISL"),
	}
}

func (h jwtHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	claim := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	tokenStr, err := token.SignedString(h.atKey)
	if err != nil {
		return err
	}

	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h jwtHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	claim := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 20)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
	tokenStr, err := token.SignedString(h.rtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
