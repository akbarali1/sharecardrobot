package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"share_card_robot/database/models"
	"share_card_robot/services"
	"share_card_robot/services/request"
)

var BotToken string
var BotUsername string

const telegramAPITimeout = 30 * time.Second

var httpClient = &http.Client{}

func TelegramAPIBase() string {
	if v := os.Getenv("TELEGRAM_API_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "https://api.telegram.org"
}

func BotInit(botToken string, botUsername string) *CommandRouter {
	BotToken = botToken
	BotUsername = strings.TrimSpace(botUsername)
	return &CommandRouter{
		commands:        make(map[string]func(*Update)),
		callbackQuery:   make(map[string]func(*CallbackQuery)),
		states:          make(map[string]func(*Update)),
		middlewareChain: &MiddlewareChain{},
	}
}

func applyMiddlewares[T any, MW ~func(func(T)) func(T)](handler func(T), mws []MW) func(T) {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func (bot *CommandRouter) HandleCommand(endpoint string, handler func(*Update), mws ...UpdateMiddleware) {
	bot.commands[endpoint] = applyMiddlewares(handler, mws)
}

func (bot *CommandRouter) HandleCallback(endpoint string, handler func(*CallbackQuery), mws ...CallbackMiddleware) {
	bot.callbackQuery[endpoint] = applyMiddlewares(handler, mws)
}

func (bot *CommandRouter) HandleState(state string, handler func(*Update), mws ...UpdateMiddleware) {
	bot.states[state] = applyMiddlewares(handler, mws)
}

func (bot *CommandRouter) HandleInlineQuery(handler func(*InlineQuery), mws ...InlineMiddleware) {
	bot.inlineQuery = applyMiddlewares(handler, mws)
}

func (bot *CommandRouter) HandleChosenInlineResult(handler func(*ChosenInlineResult), mws ...ChosenInlineResultMiddleware) {
	bot.chosenInlineResult = applyMiddlewares(handler, mws)
}

func (bot *CommandRouter) OnQuery(endpoint string, handler func(*Update), mws ...UpdateMiddleware) {
	bot.HandleCommand(endpoint, handler, mws...)
}

func decodeTelegramResponse[T any](body []byte) (T, error) {
	var response T
	err := json.Unmarshal(body, &response)
	return response, err
}

type telegramAPIResponseEnvelope struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type TelegramAPIError struct {
	Method      string
	ErrorCode   int
	Description string
	Response    string
}

func (e *TelegramAPIError) Error() string {
	if e == nil {
		return ""
	}
	if e.ErrorCode > 0 {
		return fmt.Sprintf("telegram %s failed (%d): %s", e.Method, e.ErrorCode, e.Description)
	}
	return fmt.Sprintf("telegram %s failed: %s", e.Method, e.Description)
}

func validateTelegramResponse(method string, requestPayload []byte, body []byte) error {
	var envelope telegramAPIResponseEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("telegram %s response decode: %w", method, err)
	}
	if envelope.Ok {
		return nil
	}

	log.Printf(
		"telegram request failed method=%s payload_size=%d error_code=%d description=%q",
		method,
		len(requestPayload),
		envelope.ErrorCode,
		envelope.Description,
	)

	return &TelegramAPIError{
		Method:      method,
		ErrorCode:   envelope.ErrorCode,
		Description: envelope.Description,
		Response:    string(body),
	}
}

func sendTelegram(method string, data map[string]interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/bot%s/%s", TelegramAPIBase(), BotToken, method)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("telegram %s request marshal: %w", method, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), telegramAPITimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("telegram %s build request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram %s http request: %w", method, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram %s response read: %w", method, err)
	}

	if err := validateTelegramResponse(method, jsonData, body); err != nil {
		return nil, err
	}

	return body, nil
}

func sendMessage(chatID int64, text string, opts ...map[string]interface{}) (TelegramResponse, error) {
	data := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	if len(opts) > 0 {
		for k, v := range opts[0] {
			data[k] = v
		}
	}

	prepareReplyMarkup(data)
	body, err := sendTelegram("sendMessage", data)
	if err != nil {
		return TelegramResponse{}, err
	}

	response, err := decodeTelegramResponse[TelegramResponse](body)
	if err != nil {
		return TelegramResponse{}, fmt.Errorf("telegram sendMessage decode: %w", err)
	}

	return response, nil
}

func SendMessage(chatID int64, text string, opts ...map[string]interface{}) (TelegramResponse, error) {
	return sendMessage(chatID, text, opts...)
}

func (u *Update) SendMessage(text string, opts ...map[string]interface{}) (TelegramResponse, error) {
	if u.Message == nil || u.Message.Chat == nil {
		return TelegramResponse{}, errors.New("send message: update message is nil")
	}
	return sendMessage(u.Message.Chat.ID, text, opts...)
}

func (u *Update) ReplyMessage(text string, opts ...map[string]interface{}) (TelegramResponse, error) {
	if u.Message == nil || u.Message.Chat == nil {
		return TelegramResponse{}, errors.New("reply message: update message is nil")
	}

	merged := map[string]interface{}{
		"reply_to_message_id": u.Message.MessageID,
	}
	if len(opts) > 0 {
		for k, v := range opts[0] {
			merged[k] = v
		}
	}
	return sendMessage(u.Message.Chat.ID, text, merged)
}

func DeleteMessage(chatID int64, messageID int) error {
	_, err := sendTelegram("deleteMessage", map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	})
	return err
}

func (u *Update) DeleteMessage(messageIDs ...int) error {
	if u.Message == nil || u.Message.Chat == nil {
		return errors.New("delete message: update message is nil")
	}

	messageID := u.Message.MessageID
	if len(messageIDs) > 0 {
		messageID = messageIDs[0]
	}

	return DeleteMessage(u.Message.Chat.ID, messageID)
}

func AnswerInlineQuery(inlineQueryID string, results []InlineQueryResultArticle, opts ...map[string]interface{}) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"inline_query_id": inlineQueryID,
		"results":         results,
	}

	if len(opts) > 0 {
		for k, v := range opts[0] {
			data[k] = v
		}
	}

	prepareReplyMarkup(data)
	body, err := sendTelegram("answerInlineQuery", data)
	if err != nil {
		return nil, err
	}

	response, err := decodeTelegramResponse[map[string]interface{}](body)
	if err != nil {
		return nil, fmt.Errorf("telegram answerInlineQuery decode: %w", err)
	}
	return response, nil
}

func SetMyCommands(chatID int64, cmds []BotCommand) error {
	_, err := sendTelegram("setMyCommands", map[string]interface{}{
		"commands": cmds,
		"scope": map[string]interface{}{
			"type":    "chat",
			"chat_id": chatID,
		},
	})
	return err
}

func attachAuthUser(update *Update, user *models.User) {
	if user == nil {
		return
	}

	update.AuthUser = user
	if update.InlineQuery != nil {
		update.InlineQuery.AuthUser = user
	}
	if update.CallbackQuery != nil {
		update.CallbackQuery.AuthUser = user
	}
	if update.ChosenInlineResult != nil {
		update.ChosenInlineResult.AuthUser = user
	}
}

func (bot *CommandRouter) Dispatch(update *Update) {
	if update == nil {
		return
	}

	if update.InlineQuery != nil {
		if bot.inlineQuery != nil {
			bot.inlineQuery(update.InlineQuery)
		}
		return
	}

	if update.ChosenInlineResult != nil {
		if bot.chosenInlineResult != nil {
			bot.chosenInlineResult(update.ChosenInlineResult)
		}
		return
	}

	if update.Message != nil && strings.TrimSpace(update.Message.Text) != "" {
		text := strings.TrimSpace(update.Message.Text)
		chatID := int64(0)
		if update.Message.Chat != nil {
			chatID = update.Message.Chat.ID
		}
		state := ""
		if chatID != 0 {
			state = request.GetUserStateByChatId(chatID)
		}
		clearState := func() {
			if chatID != 0 && state != "" {
				request.ClearUserStateByChatId(chatID)
				state = ""
			}
		}

		if handler, ok := bot.commands[text]; ok {
			clearState()
			handler(update)
			return
		}
		cmd := strings.SplitN(text, " ", 2)[0]
		if handler, ok := bot.commands[cmd]; ok {
			clearState()
			handler(update)
			return
		}

		if state != "" {
			if strings.HasPrefix(text, "/") {
				clearState()
			} else if handler, ok := bot.states[state]; ok {
				handler(update)
				return
			}
		}

		if handler, ok := bot.commands[services.OnQuery]; ok {
			handler(update)
			return
		}
	}

	if cb := update.CallbackQuery; cb != nil && cb.ID != "" {
		var callbackData CallbackData
		if err := json.Unmarshal([]byte(cb.Data), &callbackData); err != nil {
			log.Printf("callback data parse error: %v", err)
			return
		}

		if handler, ok := bot.callbackQuery[callbackData.Type]; ok {
			cb.UpdateID = update.UpdateID
			handler(cb)
		}
	}
}

func CreateWebhookHandler(router *CommandRouter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := request.GetParsedBody(r)
		if body == nil {
			http.Error(w, "JSON xato", http.StatusBadRequest)
			return
		}

		jsonBytes, err := json.Marshal(body)
		if err != nil {
			http.Error(w, "JSON xato", http.StatusBadRequest)
			return
		}

		var update Update
		if err := json.Unmarshal(jsonBytes, &update); err != nil {
			http.Error(w, "JSON xato", http.StatusBadRequest)
			return
		}

		if currentUser, ok := request.GetCurrentUser(r).(*models.User); ok {
			attachAuthUser(&update, currentUser)
		}

		w.WriteHeader(http.StatusOK)
		go dispatchAsync(router, update)
	}
}

func dispatchAsync(router *CommandRouter, update Update) {
	router.Dispatch(&update)
}

var allowedUpdates = []string{
	"message",
	"inline_query",
	"chosen_inline_result",
	"callback_query",
}

func SetWebhook(webhookURL string) error {
	data := map[string]interface{}{
		"url":             webhookURL,
		"allowed_updates": allowedUpdates,
	}
	if secret := os.Getenv("TELEGRAM_WEBHOOK_SECRET"); secret != "" {
		data["secret_token"] = secret
	}

	body, err := sendTelegram("setWebhook", data)
	if err != nil {
		return err
	}
	return validateTelegramResponse("setWebhook", nil, body)
}

func (bot *CommandRouter) StartServer() error {
	http.HandleFunc("/webhook", WebhookSecretMiddleware(bot.ThenMiddleware(CreateWebhookHandler(bot))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8011"
	}

	if webhookURL := strings.TrimSpace(os.Getenv("WEBHOOK_URL")); webhookURL != "" {
		if err := SetWebhook(strings.TrimRight(webhookURL, "/") + "/webhook"); err != nil {
			log.Printf("setWebhook error: %v", err)
		} else {
			log.Printf("Webhook o'rnatildi: %s/webhook", strings.TrimRight(webhookURL, "/"))
		}
	}

	log.Printf("Server %s-portda ishlamoqda. Webhook path: /webhook", port)
	return http.ListenAndServe(":"+port, nil)
}
