package callback_query

import (
	"encoding/json"
	"log"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/utils"
)

func DeleteCardPromptCallback(cb *core.CallbackQuery) {
	var callbackData core.CallbackData
	if err := jsonUnmarshalCallback(cb.Data, &callbackData); err != nil {
		_, _ = cb.AnswerCallbackQuery("❌ Noto'g'ri so'rov", map[string]interface{}{"show_alert": true})
		return
	}

	card, err := models.GetCardByIDAndUser(int64(callbackData.ID), cb.AuthUser.ID)
	if err != nil {
		log.Printf("get card for delete prompt error: %v", err)
		_, _ = cb.AnswerCallbackQuery("❌ Xatolik yuz berdi", map[string]interface{}{"show_alert": true})
		return
	}
	if card == nil {
		_, _ = cb.AnswerCallbackQuery("❌ Karta topilmadi", map[string]interface{}{"show_alert": true})
		return
	}

	_, _ = cb.AnswerCallbackQuery("⚠️ Tasdiqlang")
	_, _ = cb.BaseEditMessageText(cb.Message.MessageID, utils.RenderDeletePrompt(card), map[string]interface{}{
		"reply_markup": utils.BuildDeletePromptMarkup(card.ID, callbackData.Page),
	})
}

func DeleteCardCancelCallback(cb *core.CallbackQuery) {
	var callbackData core.CallbackData
	if err := jsonUnmarshalCallback(cb.Data, &callbackData); err != nil {
		_, _ = cb.AnswerCallbackQuery("❌ Noto'g'ri so'rov", map[string]interface{}{"show_alert": true})
		return
	}

	_, _ = cb.AnswerCallbackQuery("Bekor qilindi")
	_ = renderCardsMessage(cb, callbackData.Page)
}

func renderCardsMessage(cb *core.CallbackQuery, page int) error {
	if page <= 0 {
		page = 1
	}

	const perPage = 10

	totalCards, err := models.CountCardsByUser(cb.AuthUser.ID)
	if err != nil {
		log.Printf("list cards for callback error: %v", err)
		_, _ = cb.BaseEditMessageText(cb.Message.MessageID, "❌ Kartalarni olishda xatolik yuz berdi.")
		return err
	}

	totalPages := (totalCards + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	cards, err := models.ListCardsByUserPage(cb.AuthUser.ID, perPage, (page-1)*perPage)
	if err != nil {
		log.Printf("list cards for callback error: %v", err)
		_, _ = cb.BaseEditMessageText(cb.Message.MessageID, "❌ Kartalarni olishda xatolik yuz berdi.")
		return err
	}

	text := utils.RenderCardsList(cards, page, totalPages)
	markup := utils.BuildCardsListMarkup(cards, page, totalPages)
	if markup != nil {
		_, _ = cb.BaseEditMessageText(cb.Message.MessageID, text, map[string]interface{}{
			"reply_markup": markup,
		})
		return nil
	}
	_, _ = cb.BaseEditMessageText(cb.Message.MessageID, text)
	return nil
}

func jsonUnmarshalCallback(data string, dst any) error {
	return json.Unmarshal([]byte(data), dst)
}
