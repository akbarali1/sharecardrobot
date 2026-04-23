package request

import (
	"context"
	"net/http"
	"sync"
)

var currentRequest map[string]interface{}
var RequestMutex sync.Mutex

type parsedRequestBodyKey struct{}
type currentUserKey struct{}

var (
	userStates    = make(map[int64]string)
	userStateData = make(map[int64]map[string]string)
	stateMutex    sync.RWMutex
)

func SetRequest(data map[string]interface{}) {
	currentRequest = data
}

func GetRequest() map[string]interface{} {
	return currentRequest
}

func ClearRequest() {
	currentRequest = nil
}

func WithParsedBody(r *http.Request, body map[string]interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), parsedRequestBodyKey{}, body)
	return r.WithContext(ctx)
}

func GetParsedBody(r *http.Request) map[string]interface{} {
	if r == nil {
		return nil
	}
	if body, ok := r.Context().Value(parsedRequestBodyKey{}).(map[string]interface{}); ok {
		return body
	}
	return nil
}

func WithCurrentUser(r *http.Request, user any) *http.Request {
	ctx := context.WithValue(r.Context(), currentUserKey{}, user)
	return r.WithContext(ctx)
}

func GetCurrentUser(r *http.Request) any {
	if r == nil {
		return nil
	}
	return r.Context().Value(currentUserKey{})
}

func SetUserStateByChatId(chatID int64, state string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	userStates[chatID] = state
}

func GetUserStateByChatId(chatID int64) string {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	return userStates[chatID]
}

func SetUserDataByChatId(chatID int64, key string, value string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	if _, ok := userStateData[chatID]; !ok {
		userStateData[chatID] = make(map[string]string)
	}
	userStateData[chatID][key] = value
}

func GetUserDataByChatId(chatID int64, key string) string {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	if values, ok := userStateData[chatID]; ok {
		return values[key]
	}
	return ""
}

func ClearUserStateByChatId(chatID int64) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	delete(userStates, chatID)
	delete(userStateData, chatID)
}

func HasUserState(chatID int64) bool {
	stateMutex.RLock()
	defer stateMutex.RUnlock()
	_, exists := userStates[chatID]
	return exists
}
