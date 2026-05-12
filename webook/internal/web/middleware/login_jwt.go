package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct {
	paths []string
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	UserAgent string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{}
}

func (l *LoginJWTMiddlewareBuilder) IgnorePaths(path string) *LoginJWTMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginJWTMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if slices.Contains(l.paths, ctx.Request.URL.Path) {
			return
		}

		tokenStr := ExtractToken(ctx)

		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte("EP4UNRkorqfjiac2bt6CVH1QuCEYlISP"), nil
		})

		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token == nil || !token.Valid || claims.Uid == 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.UserAgent != ctx.Request.UserAgent() {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("claims", claims)
	}
}

func ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")

	segs := strings.SplitN(tokenHeader, " ", 2)
	if len(segs) != 2 {
		return ""
	}

	return segs[1]
}
