package core

import (
	"net/http"

	"share_card_robot/database/models"
)

type Update struct {
	UpdateID           int                 `json:"update_id"`
	AuthUser           *models.User        `json:"-"`
	Message            *Message            `json:"message,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	InlineQuery        *InlineQuery        `json:"inline_query,omitempty"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
}

type InlineQuery struct {
	AuthUser *models.User `json:"-"`
	ID       string       `json:"id"`
	From     *User        `json:"from"`
	Query    string       `json:"query"`
	Offset   string       `json:"offset"`
}

type ChosenInlineResult struct {
	AuthUser        *models.User `json:"-"`
	ResultID        string       `json:"result_id"`
	From            *User        `json:"from"`
	Query           string       `json:"query"`
	InlineMessageID string       `json:"inline_message_id,omitempty"`
}

type CallbackQuery struct {
	AuthUser *models.User `json:"-"`
	ID       string       `json:"id"`
	From     *User        `json:"from"`
	Message  *Message     `json:"message"`
	Data     string       `json:"data"`
	UpdateID int          `json:"-"`
}

type Message struct {
	MessageID       int                   `json:"message_id"`
	From            *User                 `json:"from,omitempty"`
	Chat            *Chat                 `json:"chat"`
	Text            string                `json:"text,omitempty"`
	ReplyMarkup     *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
	IsTopicMessage  bool                  `json:"is_topic_message,omitempty"`
	MessageThreadID int                   `json:"message_thread_id,omitempty"`
}

type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

type Chat struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Type      string `json:"type"`
}

type InlineQueryResultArticle struct {
	Type                string                   `json:"type"`
	ID                  string                   `json:"id"`
	Title               string                   `json:"title"`
	URL                 string                   `json:"url,omitempty"`
	Description         string                   `json:"description,omitempty"`
	ThumbnailURL        string                   `json:"thumbnail_url,omitempty"`
	ThumbnailWidth      int                      `json:"thumbnail_width,omitempty"`
	ThumbnailHeight     int                      `json:"thumbnail_height,omitempty"`
	InputMessageContent *InputTextMessageContent `json:"input_message_content"`
	ReplyMarkup         *InlineKeyboardMarkup    `json:"reply_markup,omitempty"`
}

type InputTextMessageContent struct {
	MessageText string `json:"message_text"`
	ParseMode   string `json:"parse_mode,omitempty"`
}

type KeyboardButton struct {
	Text                         string       `json:"text"`
	CallbackData                 CallbackData `json:"-"`
	Data                         string       `json:"callback_data,omitempty"`
	URL                          string       `json:"url,omitempty"`
	SwitchInlineQueryCurrentChat *string      `json:"switch_inline_query_current_chat,omitempty"`
	SwitchInlineQuery            *string      `json:"switch_inline_query,omitempty"`
}

type CallbackData struct {
	Type string `json:"t"`
	ID   int    `json:"id"`
	Page int    `json:"p,omitempty"`
}

type ReplyKeyboardButton struct {
	Text string `json:"text"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        [][]ReplyKeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool                    `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard bool                    `json:"one_time_keyboard,omitempty"`
	Selective       bool                    `json:"selective,omitempty"`
}

type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
	Selective      bool `json:"selective,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard  [][]KeyboardButton `json:"inline_keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard bool               `json:"one_time_keyboard,omitempty"`
	Selective       bool               `json:"selective,omitempty"`
	Message         string             `json:"-"`
}

type TelegramResponse struct {
	Ok     bool     `json:"ok"`
	Result *Message `json:"result,omitempty"`
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

type MiddlewareChain struct {
	middlewares []Middleware
}

type UpdateMiddleware func(func(*Update)) func(*Update)
type CallbackMiddleware func(func(*CallbackQuery)) func(*CallbackQuery)
type InlineMiddleware func(func(*InlineQuery)) func(*InlineQuery)
type ChosenInlineResultMiddleware func(func(*ChosenInlineResult)) func(*ChosenInlineResult)

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type CommandRouter struct {
	commands           map[string]func(*Update)
	callbackQuery      map[string]func(*CallbackQuery)
	states             map[string]func(*Update)
	inlineQuery        func(*InlineQuery)
	chosenInlineResult func(*ChosenInlineResult)
	middlewareChain    *MiddlewareChain
}
