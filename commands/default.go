package commands

import (
	"strings"

	"share_card_robot/core"
)

func DefaultHandler(ctx *core.Update) {
	if looksLikeSharedCardMessage(ctx.Message.Text) {
		_, _ = ctx.ReplyMessage("ℹ️ Bu karta boshqalarga ham shu ko'rinishda yuboriladi.")
		return
	}

	_, _ = ctx.SendMessage("❓ Buyruq tushunilmadi.\n\n" + buildHelpText())
}

func looksLikeSharedCardMessage(text string) bool {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "💳 ") {
		return false
	}

	lines := strings.Split(text, "\n")
	return len(lines) >= 2
}
