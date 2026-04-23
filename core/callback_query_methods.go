package core

import (
	"errors"
	"fmt"
)

func (c *CallbackQuery) SendMessage(text string, opts ...map[string]interface{}) (TelegramResponse, error) {
	if c.Message == nil || c.Message.Chat == nil {
		return TelegramResponse{}, errors.New("callback send message: message is nil")
	}
	return sendMessage(c.Message.Chat.ID, text, opts...)
}

func (c *CallbackQuery) DeleteMessage(messageIDs ...int) error {
	if c.Message == nil || c.Message.Chat == nil {
		return errors.New("callback delete message: message is nil")
	}

	messageID := c.Message.MessageID
	if len(messageIDs) > 0 {
		messageID = messageIDs[0]
	}

	return DeleteMessage(c.Message.Chat.ID, messageID)
}

func (c *CallbackQuery) AnswerCallbackQuery(text string, opts ...map[string]interface{}) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"callback_query_id": c.ID,
		"text":              text,
	}

	if len(opts) > 0 {
		for k, v := range opts[0] {
			data[k] = v
		}
	}

	body, err := sendTelegram("answerCallbackQuery", data)
	if err != nil {
		return nil, err
	}

	response, err := decodeTelegramResponse[map[string]interface{}](body)
	if err != nil {
		return nil, fmt.Errorf("telegram answerCallbackQuery decode: %w", err)
	}

	return response, nil
}

func (c *CallbackQuery) BaseEditMessageText(messageID int, text string, opts ...map[string]interface{}) (map[string]interface{}, error) {
	if c.Message == nil || c.Message.Chat == nil {
		return nil, errors.New("callback edit message text: message is nil")
	}

	data := map[string]interface{}{
		"chat_id":    c.Message.Chat.ID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}

	if len(opts) > 0 {
		for k, v := range opts[0] {
			data[k] = v
		}
	}

	prepareReplyMarkup(data)
	body, err := sendTelegram("editMessageText", data)
	if err != nil {
		return nil, err
	}

	response, err := decodeTelegramResponse[map[string]interface{}](body)
	if err != nil {
		return nil, fmt.Errorf("telegram editMessageText decode: %w", err)
	}

	return response, nil
}
