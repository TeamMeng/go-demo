package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockLimiter struct {
	limited bool
	err     error
}

func (m *mockLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return m.limited, m.err
}

func TestBuilder_Build(t *testing.T) {
	const limitURL = "/limit"
	tests := []struct {
		name string

		limited    bool
		limiterErr error
		reqBuilder func(t *testing.T) *http.Request

		wantCode int
	}{
		{
			name:       "不限流",
			limited:    false,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusOK,
		},
		{
			name:       "限流",
			limited:    true,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusTooManyRequests,
		},
		{
			name:       "系统错误",
			limited:    false,
			limiterErr: errors.New("模拟系统错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, limitURL, nil)
				if err != nil {
					t.Fatal(err)
				}
				return req
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewBuilder(&mockLimiter{limited: tt.limited, err: tt.limiterErr})

			server := gin.Default()
			server.Use(svc.Build())
			svc.RegisterRoutes(server)

			req := tt.reqBuilder(t)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)

			assert.Equal(t, tt.wantCode, recorder.Code)
		})
	}
}

func TestBuilder_limit(t *testing.T) {
	tests := []struct {
		name string

		limited    bool
		limiterErr error
		reqBuilder func(t *testing.T) *http.Request

		want    bool
		wantErr error
	}{
		{
			name:       "不限流",
			limited:    false,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: false,
		},
		{
			name:       "限流",
			limited:    true,
			limiterErr: nil,
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want: true,
		},
		{
			name:       "限流代码出错",
			limited:    false,
			limiterErr: errors.New("模拟系统错误"),
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, "", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.RemoteAddr = "127.0.0.1:80"
				return req
			},
			want:    false,
			wantErr: errors.New("模拟系统错误"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(&mockLimiter{limited: tt.limited, err: tt.limiterErr})

			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			req := tt.reqBuilder(t)
			ctx.Request = req

			got, err := b.limit(ctx)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func (b *Builder) RegisterRoutes(server *gin.Engine) {
	server.GET("/limit", func(ctx *gin.Context) {
		ctx.Status(http.StatusOK)
	})
}
