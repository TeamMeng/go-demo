package ioc

import (
	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
	"github.com/TeamMeng/go-demo/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	svc := memory.NewService()
	return svc
}
