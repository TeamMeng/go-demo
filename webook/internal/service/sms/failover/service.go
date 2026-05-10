package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
	idx  uint64
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

	return errors.New("all sms failed")
}

func (f *FailoverSMSService) SendV1(ctx context.Context, biz string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))

	for i := idx; i < idx+length; i++ {
		svc := f.svcs[i%length]
		err := svc.Send(ctx, biz, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		}
	}

	return errors.New("all sms failed")
}
