package auth

import (
	"context"
	"errors"

	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key string
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}

func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var tc Claims
	token, err := jwt.ParseWithClaims(biz, &tc, func(t *jwt.Token) (any, error) {
		return s.key, nil
	})

	if err != nil {
		return nil
	}

	if !token.Valid {
		return errors.New("token invalid")
	}

	return s.svc.Send(ctx, biz, args, numbers...)
}
