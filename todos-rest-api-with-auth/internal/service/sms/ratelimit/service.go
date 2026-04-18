package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"todo_api/internal/ratelimit"
	"todo_api/internal/service/sms"
)

type RateLimitSmsService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(limiter ratelimit.Limiter, svc sms.Service) *RateLimitSmsService {
	return &RateLimitSmsService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RateLimitSmsService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms")
	if err != nil {
		return fmt.Errorf("短信限流异常: %w", err)
	}
	if limited {
		return errors.New("触发限流")
	}
	return s.svc.Send(ctx, tpl, args, numbers...)
}
