package commands

import (
	"fmt"

	"share_card_robot/config"
	"share_card_robot/core"
)

func AboutHandler(ctx *core.Update) {
	text := fmt.Sprintf(`ℹ️ <b>Mening Kartalarim</b> — v%s

👨‍💻 <b>Yaratuvchi:</b> Akbarali

🌐 <b>Ochiq manba (Open Source):</b>
Loyiha butunlay ochiq — kodni ko'rish, taklif kiritish yoki o'z serveringizda ishlatishingiz mumkin.

🔗 <a href="https://github.com/akbarali1/share_card_robot">github.com/akbarali1/share_card_robot</a>`, config.Version)

	_, _ = ctx.SendMessage(text, map[string]interface{}{
		"disable_web_page_preview": true,
	})
}
