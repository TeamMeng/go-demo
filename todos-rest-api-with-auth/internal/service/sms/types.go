package sms

import "context"

type Service interface {
	Send(ctx context.Context, blz string, args []string, numbers ...string) error
}
