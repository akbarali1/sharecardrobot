package core

import "share_card_robot/services/request"

func SetUserState(chatID int64, state string) {
	request.SetUserStateByChatId(chatID, state)
}

func GetUserState(chatID int64) string {
	return request.GetUserStateByChatId(chatID)
}

func ClearUserState(chatID int64) {
	request.ClearUserStateByChatId(chatID)
}

func HasUserState(chatID int64) bool {
	return request.HasUserState(chatID)
}

func (u *Update) SetState(state string) {
	if u.Message != nil && u.Message.Chat != nil {
		SetUserState(u.Message.Chat.ID, state)
	}
}

func (u *Update) GetState() string {
	if u.Message != nil && u.Message.Chat != nil {
		return GetUserState(u.Message.Chat.ID)
	}
	return ""
}

func (u *Update) ClearState() {
	if u.Message != nil && u.Message.Chat != nil {
		ClearUserState(u.Message.Chat.ID)
	}
}

func (u *Update) HasState() bool {
	if u.Message != nil && u.Message.Chat != nil {
		return HasUserState(u.Message.Chat.ID)
	}
	return false
}
