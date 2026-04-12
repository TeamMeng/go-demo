package service

import (
	"context"
	"fmt"
	"math/rand"

	"todo_api/internal/repository"
	"todo_api/internal/service/sms"
)

// codeTplId 是短信服务商侧配置的验证码短信模板 ID。
//
// 当前值为空，表示还没有绑定真实模板；接入真实短信通道时需要替换成
// 阿里云、腾讯云等短信平台审核通过的模板编号。
const codeTplId = "100001"

// CodeService 负责验证码业务流程编排。
//
// 它不直接操作 Redis，也不直接拼装短信服务商的底层请求，而是把职责拆给：
//  1. CodeRepository：保存验证码、校验验证码、控制发送和验证频率。
//  2. sms.Service：把生成好的验证码发送到用户手机。
//
// 这样 service 层只关注“先生成并存储验证码，再发送短信”这条业务流程。
type CodeService struct {
	// repo 封装验证码的持久化和校验逻辑，目前底层实现是 Redis。
	repo *repository.CodeRepository
	// smsSvc 是短信服务抽象，可以替换成腾讯云、阿里云或测试用 mock 实现。
	smsSvc sms.Service
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

// Send 生成验证码、写入缓存并发送短信。
//
// 参数说明：
//   - biz：业务场景，例如 login、register、reset_password。不同业务场景会写入不同 Redis key，
//     避免登录验证码和注册验证码互相覆盖。
//   - phone：接收验证码的手机号。
//
// 执行顺序：
//  1. 生成 6 位验证码。
//  2. 先写入 Redis，并由 repository/cache 层控制同一手机号短时间内不能频繁发送。
//  3. Redis 写入成功后再调用短信服务发送，避免用户收到无法校验的验证码。
//
// 如果 Redis 限流或短信发送失败，会直接返回对应错误给调用方。
func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	// 生产一个验证码
	code := svc.generateCode()
	// 放入redis
	if err := svc.repo.Store(ctx, biz, phone, code); err != nil {
		return err
	}
	// 发送出去
	if err := svc.smsSvc.Send(ctx, codeTplId, []string{code, "5"}, phone); err != nil {
		return err
	}

	return nil
}

// Verify 校验用户输入的验证码。
//
// 参数说明：
//   - biz：必须和 Send 时传入的业务场景一致，否则会查到不同的验证码 key。
//   - phone：验证码归属的手机号。
//   - inputCode：用户提交的验证码。
//
// 返回值说明：
//   - bool 为 true：验证码正确，且底层会把可验证次数置为不可再次验证，防止重复使用。
//   - bool 为 false 且 error 为 nil：验证码错误，但仍属于正常业务结果。
//   - error 不为 nil：验证码不存在、验证次数耗尽、Redis 异常等错误。
func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

// generateCode 生成 6 位数字验证码。
//
// rand.Intn(1000000) 的取值范围是 [0, 999999]，再格式化为 6 位字符串用于短信发送。
// 注意：这里使用的是 math/rand，适合普通业务验证码；如果后续验证码承担更高安全要求，
// 可以考虑切换为 crypto/rand。
func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
