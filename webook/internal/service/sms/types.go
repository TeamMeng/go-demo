package sms

import "context"

type Service interface {
	Send(ctx context.Context, tpl string, atgs []string, numbers ...string) error
}
