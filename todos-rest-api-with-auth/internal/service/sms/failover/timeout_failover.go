package failover

import (
	"context"
	"sync/atomic"
	"todo_api/internal/service/sms"
)

type timeoutFailover struct {
	svcs      []sms.Service
	idx       int32
	cnt       int32
	threshold int32
}

func NewTimeoutFailoverSMSService() sms.Service {
	return &timeoutFailover{}
}

func (t *timeoutFailover) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
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
	err := svc.Send(ctx, biz, args, numbers...)

	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
		return err
	case nil:
		atomic.StoreInt32(&t.cnt, 0)
		return err
	default:
		return err
	}
}
