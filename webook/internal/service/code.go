package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/TeamMeng/go-demo/webook/internal/repository"
	"github.com/TeamMeng/go-demo/webook/internal/service/sms"
)

const codeTplId = "100001"

type CodeService struct {
	repo   *repository.CodeRepository
	smsSvc sms.Service
}

func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generateCode()
	if err := svc.repo.Store(ctx, biz, phone, code); err != nil {
		return err
	}

	return svc.smsSvc.Send(ctx, codeTplId, []string{code, "5"}, phone)
}

func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
