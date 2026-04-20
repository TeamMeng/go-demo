package failover

import (
	"context"
	"errors"
	"log"
	"todo_api/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, biz, args, numbers...)
		if err == nil {
			return nil
		}
		log.Println(err)
	}
	return errors.New("All SMS failed")
}

func (f *FailoverSMSService) SendV1(ctx context.Context, biz string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, biz, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		default:
			log.Println(err)
		}
	}
	return errors.New("All SMS failed")
}
