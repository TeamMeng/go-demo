package middleware

import (
	"encoding/gob"
	"net/http"
	"slices"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		if slices.Contains(l.paths, ctx.Request.URL.Path) {
			return
		}

		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		updateTime := sess.Get("update_time")
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		now := time.Now()
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
		}
		updateTimeVal, ok := updateTime.(time.Time)

		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if now.Sub(updateTimeVal) > time.Minute {
			sess.Set("update_time", now)
			sess.Save()
		}
	}
}
