package aliyun

import (
	"context"
	"encoding/json"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

// Service 是阿里云短信服务实现。
//
// 它实现了 internal/service/sms.Service 接口，业务层只依赖接口，不需要知道底层使用的是
// 阿里云 dysmsapi SDK。
type Service struct {
	// signName 是阿里云短信签名名称，必须是阿里云控制台审核通过的签名。
	signName *string
	// client 是阿里云短信 SDK 客户端，负责发起 SendSms 请求。
	client *dysmsapi.Client
}

// NewService 创建阿里云短信服务。
//
// signName 和 client 由外部传入，便于在启动阶段统一初始化 SDK 客户端，也便于测试时替换。
func NewService(signName *string, client *dysmsapi.Client) *Service {
	return &Service{
		client:   client,
		signName: signName,
	}
}

// CreateClient 基于阿里云默认凭证链创建短信客户端。
//
// credential.NewCredential(nil) 会按阿里云 SDK 的默认规则查找凭证，例如环境变量、
// 配置文件、实例角色等。这里固定使用 dysmsapi.aliyuncs.com 作为短信服务 endpoint。
func CreateClient() (*dysmsapi.Client, error) {
	cred, err := credential.NewCredential(nil)
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		Credential: cred,
	}
	config.Endpoint = tea.String("dysmsapi.aliyuncs.com")

	client, err := dysmsapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Send 发送短信。
//
// 参数说明：
//   - tpl：阿里云短信模板 Code。
//   - args：模板参数值，会按顺序映射到 code、min、time、name、product。
//     超过预置参数名数量的值，会映射为 param{index}，例如 param5。
//   - numbers：接收短信的手机号列表。
//
// 阿里云 SendSms 一次请求只发送给一个 PhoneNumbers 字段；这里对 numbers 做循环，
// 逐个构造请求并发送。只要任意一个号码发送失败，就立即返回错误。
func (s *Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	// paramMap 用来生成阿里云 TemplateParam JSON。
	// 例如 args = ["123456"] 时，会生成 {"code":"123456"}。
	paramMap := make(map[string]string)
	// keys 是常见短信模板参数名的约定顺序；业务层只传参数值，短信实现负责映射为服务商格式。
	keys := []string{"code", "min", "time", "name", "product"}
	for i, arg := range args {
		if i < len(keys) {
			paramMap[keys[i]] = arg
		} else {
			paramMap[fmt.Sprintf("param%d", i)] = arg
		}
	}

	paramJSON, err := json.Marshal(paramMap)
	if err != nil {
		return fmt.Errorf("failed to marshal template params: %w", err)
	}

	// 遍历每个号码发送
	for _, number := range numbers {
		// SendSmsRequest 中的 TemplateParam 必须是 JSON 字符串，SignName 和 TemplateCode
		// 必须和阿里云控制台审核通过的短信签名、模板保持一致。
		req := &dysmsapi.SendSmsRequest{
			PhoneNumbers:  tea.String(number),
			SignName:      s.signName,
			TemplateCode:  tea.String(tpl),
			TemplateParam: tea.String(string(paramJSON)),
		}

		_, err := s.client.SendSmsWithOptions(req, nil)
		if err != nil {
			return fmt.Errorf("failed to send SMS to %s: %w", number, err)
		}
	}
	return nil
}
