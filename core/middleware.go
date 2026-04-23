package core

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"share_card_robot/database/models"
	"share_card_robot/services/request"
)

func (bot *CommandRouter) AddMiddleware(mw Middleware) {
	bot.middlewareChain.middlewares = append(bot.middlewareChain.middlewares, mw)
}

func (bot *CommandRouter) ThenMiddleware(h http.HandlerFunc) http.HandlerFunc {
	for i := len(bot.middlewareChain.middlewares) - 1; i >= 0; i-- {
		h = bot.middlewareChain.middlewares[i](h)
	}
	return h
}

func SetRequest(w http.ResponseWriter, r *http.Request) (*http.Request, map[string]interface{}, bool) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON xato", http.StatusBadRequest)
		return r, nil, false
	}
	defer r.Body.Close()
	r = request.WithParsedBody(r, body)
	//println("Request: " + StructToString(body))
	return r, body, true
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Faqat POST so'rovlari qabul qilinadi", http.StatusMethodNotAllowed)
			return
		}

		r, body, ok := SetRequest(w, r)
		if !ok {
			return
		}

		if currentUser := refreshAuthUser(body); currentUser != nil {
			r = request.WithCurrentUser(r, currentUser)
		}

		next.ServeHTTP(w, r)
	}
}

func TimingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[Timing] %s ishlash vaqti: %v", r.URL.Path, time.Since(start))
	}
}

func WebhookSecretMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("TELEGRAM_WEBHOOK_SECRET")
		if secret != "" {
			got := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
			if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
				log.Printf("webhook: noto'g'ri secret token (ip=%s)", r.RemoteAddr)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}

func ErrorHandlingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("[Recovery] %s panic: %v\n", r.URL.Path, err)
				println(string(debug.Stack()))
				w.WriteHeader(http.StatusOK)
			}
		}()

		next.ServeHTTP(w, r)
	}
}

func refreshAuthUser(body map[string]interface{}) *models.User {
	paths := [][]string{
		{"message", "from"},
		{"callback_query", "from"},
		{"inline_query", "from"},
		{"chosen_inline_result", "from"},
	}

	for _, path := range paths {
		if user := userFromBody(body, path...); user != nil {
			return user
		}
	}

	return nil
}

func userFromBody(body map[string]interface{}, path ...string) *models.User {
	userMap, ok := getNestedMap(body, path...)
	if !ok {
		return nil
	}

	telegramUserID, ok := getInt64(userMap["id"])
	if !ok || telegramUserID == 0 {
		return nil
	}

	firstName, _ := userMap["first_name"].(string)
	lastName, _ := userMap["last_name"].(string)
	username, _ := userMap["username"].(string)

	user, err := models.GetOrCreateTelegramUser(
		telegramUserID,
		strings.TrimSpace(username),
		strings.TrimSpace(firstName),
		strings.TrimSpace(lastName),
	)
	if err != nil {
		log.Printf("auth user sync error: %v", err)
		return nil
	}
	return user
}

func getNestedMap(body map[string]interface{}, path ...string) (map[string]interface{}, bool) {
	var current any = body
	for _, key := range path {
		next, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current = next[key]
	}

	result, ok := current.(map[string]interface{})
	return result, ok
}

func getInt64(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	default:
		return 0, false
	}
}
