package aliyun

import (
	"context"
	"encoding/json"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	dypnsapi "github.com/alibabacloud-go/dypnsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

type Service struct {
	signName *string
	client   *dypnsapi.Client
}

func NewService(signName *string, client *dypnsapi.Client) *Service {
	return &Service{
		client:   client,
		signName: signName,
	}
}

func CreateClient() (*dypnsapi.Client, error) {
	cred, err := credential.NewCredential(nil)
	if err != nil {
		return nil, err
	}

	config := &openapi.Config{
		Credential: cred,
	}
	endpoint := "dypnsapi.aliyuncs.com"
	config.Endpoint = &endpoint

	client, err := dypnsapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	paramMap := make(map[string]string)
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

	for _, number := range numbers {
		phoneNumber := number
		templateCode := biz
		templateParam := string(paramJSON)
		req := &dypnsapi.SendSmsVerifyCodeRequest{
			PhoneNumber:   &phoneNumber,
			SignName:      s.signName,
			TemplateCode:  &templateCode,
			TemplateParam: &templateParam,
		}

		resp, err := s.client.SendSmsVerifyCodeWithOptions(req, &dara.RuntimeOptions{})
		if err != nil {
			return fmt.Errorf("failed to send SMS to %s: %w", number, err)
		}
		var code, message, requestID string
		if resp != nil && resp.Body != nil {
			code = tea.StringValue(resp.Body.Code)
			message = tea.StringValue(resp.Body.Message)
			requestID = tea.StringValue(resp.Body.RequestId)
		}
		if code != "OK" {
			return fmt.Errorf("failed to send SMS to %s: aliyun code=%s message=%s requestId=%s",
				number,
				code,
				message,
				requestID,
			)
		}
	}
	return nil
}
