package retryable

import (
	"context"
	"errors"

	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
)

type Service struct {
	svc      sms.Service
	retryMax int
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numerbser ...string) error {
	err := s.svc.Send(ctx, biz, args, numerbser...)
	cnt := 1

	for err != nil && cnt < s.retryMax {
		err = s.svc.Send(ctx, biz, args, numerbser...)
		if err == nil {
			return nil
		}
		cnt++
	}

	return errors.New("Sms all retry failed")
}
