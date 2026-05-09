package failover

import (
	"context"
	"sync/atomic"

	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
)

type TimeoutFailoverSMSService struct {
	svcs []sms.Service

	idx int32

	cnt int32

	threshold int32
}

func NewTimeoutFailoverSMSService(svcs []sms.Service) sms.Service {
	return &TimeoutFailoverSMSService{
		svcs: svcs,
		cnt:  5,
	}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context,
	tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	if cnt > t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = atomic.LoadInt32(&t.idx)
	}

	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case nil:
		atomic.StoreInt32(&t.cnt, 0)
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		return err
	}

	return err
}
