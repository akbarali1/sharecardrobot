package commands

import (
	"log"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/utils"
)

func MyCardsHandler(ctx *core.Update) {
	const perPage = 8

	totalCards, err := models.CountCardsByUser(ctx.AuthUser.ID)
	if err != nil {
		log.Printf("count cards error: %v", err)
		_, _ = ctx.SendMessage("❌ Kartalarni olishda xatolik yuz berdi.")
		return
	}

	totalPages := (totalCards + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	cards, err := models.ListCardsByUserPage(ctx.AuthUser.ID, perPage, 0)
	if err != nil {
		log.Printf("list cards error: %v", err)
		_, _ = ctx.SendMessage("❌ Kartalarni olishda xatolik yuz berdi.")
		return
	}

	text := utils.RenderCardsList(cards, 1, totalPages)
	markup := utils.BuildCardsListMarkup(cards, 1, totalPages)
	if markup != nil {
		_, _ = ctx.SendMessage(text, map[string]interface{}{"reply_markup": markup})
		return
	}
	_, _ = ctx.SendMessage(text)
}
