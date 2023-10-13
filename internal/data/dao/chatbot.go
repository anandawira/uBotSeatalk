package dao

import (
	"context"
	"encoding/json"
)

type ChatBotDAO interface {
	GetListSubscriber(ctx context.Context) ([]string, error)
	SendChatToSubscriber(ctx context.Context, employeeCode string, message json.RawMessage) error
}
