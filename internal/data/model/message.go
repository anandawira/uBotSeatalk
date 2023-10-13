package model

import "encoding/json"

type Text struct {
	Content string `json:"content"`
}

type TextMessage struct {
	Tag  string `json:"tag"`
	Text Text   `json:"text"`
}

func NewTextMessage(content string) (json.RawMessage, error) {
	b, err := json.Marshal(TextMessage{
		Tag: "text",
		Text: Text{
			Content: content,
		},
	})

	if err != nil {
		return nil, err
	}

	return b, nil
}
