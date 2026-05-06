package middlewares

import (
	"log"

	"share_card_robot/core"
	"share_card_robot/database/models"
)

func RequireCurrentUserInline(next func(*core.InlineQuery)) func(*core.InlineQuery) {
	return func(ctx *core.InlineQuery) {
		if ctx == nil {
			return
		}

		if ctx.AuthUser == nil {
			answerInlineNotice(ctx.ID, "❌ Avval /start ni bosing")
			return
		}

		next(ctx)
	}
}

func RequireUserHasCardsInline(next func(*core.InlineQuery)) func(*core.InlineQuery) {
	return func(ctx *core.InlineQuery) {
		cards, err := models.ListCardsByUser(ctx.AuthUser.ID, 1)
		if err != nil {
			log.Printf("RequireUserHasCardsInline: %v", err)
			answerInlineNotice(ctx.ID, "❌ Xatolik yuz berdi")
			return
		}

		if len(cards) == 0 {
			answerInlineNotice(ctx.ID, "💳 Kartangiz yo'q. Avval karta qo'shing", "add_card")
			return
		}

		next(ctx)
	}
}

func RequireCurrentUserChosenInline(next func(*core.ChosenInlineResult)) func(*core.ChosenInlineResult) {
	return func(ctx *core.ChosenInlineResult) {
		if ctx == nil || ctx.AuthUser == nil {
			return
		}
		next(ctx)
	}
}

func answerInlineNotice(inlineID string, message string, param ...string) {
	pmParam := "start"
	if len(param) > 0 && param[0] != "" {
		pmParam = param[0]
	}
	if _, err := core.AnswerInlineQuery(inlineID, []core.InlineQueryResultArticle{}, map[string]interface{}{
		"cache_time":          0,
		"is_personal":         true,
		"switch_pm_text":      message,
		"switch_pm_parameter": pmParam,
	}); err != nil {
		log.Printf("answerInlineNotice error: %v", err)
	}
}
