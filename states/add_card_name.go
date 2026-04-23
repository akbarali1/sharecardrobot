package states

import (
	"strings"

	"share_card_robot/core"
	"share_card_robot/services/request"
)

func AddCardNameState(ctx *core.Update) {
	title := strings.TrimSpace(ctx.Message.Text)
	if title == "" {
		_, _ = ctx.SendMessage("❌ Karta nomi bo'sh bo'lmasligi kerak. Qayta yuboring.")
		return
	}

	request.SetUserDataByChatId(ctx.Message.Chat.ID, "card_title", title)
	ctx.SetState("add_card_number")
	_, _ = ctx.SendMessage("💳 Endi karta raqamini yuboring.\n\nFaqat raqam yuboring. Bo'sh joy va tire bo'lsa, bot o'zi tozalaydi.")
}
