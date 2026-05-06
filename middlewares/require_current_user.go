package middlewares

import (
	"log"

	"share_card_robot/core"
)

func RequireCurrentUser(next func(*core.Update)) func(*core.Update) {
	return func(ctx *core.Update) {
		if ctx == nil || ctx.Message == nil || ctx.Message.Chat == nil {
			log.Println("RequireCurrentUser: update message context is nil")
			return
		}

		if ctx.AuthUser == nil {
			_, _ = ctx.SendMessage("❌ Foydalanuvchi aniqlanmadi. /start ni qayta yuboring.")
			return
		}

		if ctx.Message.Chat.Type != "private" {
			_, _ = ctx.SendMessage("🔒 Bu amal faqat botning private chatida ishlaydi.")
			return
		}

		next(ctx)
	}
}

func RequireCurrentUserCallback(next func(*core.CallbackQuery)) func(*core.CallbackQuery) {
	return func(ctx *core.CallbackQuery) {
		if ctx == nil || ctx.Message == nil || ctx.Message.Chat == nil {
			log.Println("RequireCurrentUserCallback: callback message context is nil")
			return
		}

		if ctx.AuthUser == nil {
			_, _ = ctx.AnswerCallbackQuery("❌ Foydalanuvchi aniqlanmadi. /start ni qayta yuboring.", map[string]interface{}{
				"show_alert": true,
			})
			return
		}

		if ctx.Message.Chat.Type != "private" {
			_, _ = ctx.AnswerCallbackQuery("🔒 Bu amal faqat botning private chatida ishlaydi.", map[string]interface{}{
				"show_alert": true,
			})
			return
		}

		next(ctx)
	}
}
