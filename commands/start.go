package commands

import (
	"fmt"
	"strings"

	"share_card_robot/core"
	"share_card_robot/utils"
)

func StartHandler(ctx *core.Update) {
	utils.SetDefaultCommands(ctx.Message.Chat.ID)

	parts := strings.SplitN(strings.TrimSpace(ctx.Message.Text), " ", 2)
	if len(parts) == 2 && parts[1] == "add_card" {
		AddCardHandler(ctx)
		return
	}

	text := fmt.Sprintf("Salom, <b>%s</b>.\n\n%s", ctx.AuthUser.DisplayName(), buildHelpText())
	_, _ = ctx.SendMessage(text, map[string]interface{}{
		"reply_markup": utils.MainKeyboard(),
	})

	tryText := "👇 O'zingiz sinab ko'ring"
	tryMarkup := &core.InlineKeyboardMarkup{
		InlineKeyboard: [][]core.KeyboardButton{
			{core.NewSwitchCurrentChatButton("🔍 Shu chatda sinab ko'ring", "")},
			{core.NewSwitchChatButton("📤 Boshqa chatga yuborish", "")},
		},
	}
	_, _ = ctx.SendMessage(tryText, map[string]interface{}{
		"reply_markup": tryMarkup,
	})
}
