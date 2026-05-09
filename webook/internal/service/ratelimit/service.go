package ratelimit

import (
	"context"
	"errors"
	"fmt"

	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
	"github.com/TeamMeng/go-demo/webook/pkg/ratelimit"
)

var errLimited = errors.New("sms limited")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms")
	if err != nil {
		return fmt.Errorf("rate limit check failed: %w", err)
	}

	if limited {
		return errLimited
	}

	err = s.svc.Send(ctx, tpl, args, numbers...)
	return err
}
