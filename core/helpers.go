package core

import (
	"encoding/json"
	"fmt"
)

func strPtr(s string) *string { return &s }

func NewSwitchCurrentChatButton(text, query string) KeyboardButton {
	return KeyboardButton{Text: text, SwitchInlineQueryCurrentChat: strPtr(query)}
}

func NewSwitchChatButton(text, query string) KeyboardButton {
	return KeyboardButton{Text: text, SwitchInlineQuery: strPtr(query)}
}

func NewCallbackButton(text string, callbackType string, id int) KeyboardButton {
	return KeyboardButton{
		Text: text,
		CallbackData: CallbackData{
			Type: callbackType,
			ID:   id,
		},
	}
}

func NewCallbackButtonWithPage(text string, callbackType string, id int, page int) KeyboardButton {
	return KeyboardButton{
		Text: text,
		CallbackData: CallbackData{
			Type: callbackType,
			ID:   id,
			Page: page,
		},
	}
}

func normalizeInlineKeyboardMarkup(markup *InlineKeyboardMarkup) *InlineKeyboardMarkup {
	if markup == nil {
		return nil
	}

	for rowIndex := range markup.InlineKeyboard {
		for colIndex := range markup.InlineKeyboard[rowIndex] {
			button := &markup.InlineKeyboard[rowIndex][colIndex]
			if button.Data != "" || button.URL != "" || button.SwitchInlineQueryCurrentChat != nil || button.SwitchInlineQuery != nil {
				continue
			}
			marshalled, err := json.Marshal(button.CallbackData)
			if err != nil {
				continue
			}
			button.Data = string(marshalled)
		}
	}

	return markup
}

func prepareReplyMarkup(data map[string]interface{}) {
	rawMarkup, ok := data["reply_markup"]
	if !ok || rawMarkup == nil {
		return
	}

	switch markup := rawMarkup.(type) {
	case *ReplyKeyboardMarkup:
		data["reply_markup"] = markup
	case ReplyKeyboardMarkup:
		copied := markup
		data["reply_markup"] = &copied
	case *ReplyKeyboardRemove:
		data["reply_markup"] = markup
	case ReplyKeyboardRemove:
		copied := markup
		data["reply_markup"] = &copied
	case *InlineKeyboardMarkup:
		data["reply_markup"] = normalizeInlineKeyboardMarkup(markup)
	case InlineKeyboardMarkup:
		copied := markup
		data["reply_markup"] = normalizeInlineKeyboardMarkup(&copied)
	}
}

func ChunkBtnToRows(buttons []KeyboardButton, chunkSize int) [][]KeyboardButton {
	if chunkSize <= 0 {
		chunkSize = 1
	}

	var rows [][]KeyboardButton
	for i := 0; i < len(buttons); i += chunkSize {
		end := i + chunkSize
		if end > len(buttons) {
			end = len(buttons)
		}

		row := make([]KeyboardButton, 0, end-i)
		for _, button := range buttons[i:end] {
			if button.Data == "" && button.URL == "" && button.SwitchInlineQueryCurrentChat == nil && button.SwitchInlineQuery == nil {
				marshalled, err := json.Marshal(button.CallbackData)
				if err != nil {
					continue
				}
				button.Data = string(marshalled)
			}
			row = append(row, button)
		}
		rows = append(rows, row)
	}
	return rows
}

func InlineKeyboard(message string, keyboard []KeyboardButton, chunkSize int) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		Message:         message,
		InlineKeyboard:  ChunkBtnToRows(keyboard, chunkSize),
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		Selective:       true,
	}
}

func StructToString(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(body)
}
