package utils

import (
	"log"
	"share_card_robot/core"
)

func SetDefaultCommands(chatID int64) {
	cmds := []core.BotCommand{
		{Command: "start", Description: "Boshlash"},
		{Command: "my_cards", Description: "Mening kartalarim"},
		{Command: "add_card", Description: "Karta qo'shish"},
		{Command: "help", Description: "Yordam"},
		{Command: "about", Description: "Bot haqida"},
	}

	if err := core.SetMyCommands(chatID, cmds); err != nil {
		log.Println("setMyCommands error:", err)
	}
}
