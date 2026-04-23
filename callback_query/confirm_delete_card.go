package callback_query

import (
	"log"

	"share_card_robot/core"
	"share_card_robot/database/models"
)

func DeleteCardConfirmCallback(cb *core.CallbackQuery) {
	var callbackData core.CallbackData
	if err := jsonUnmarshalCallback(cb.Data, &callbackData); err != nil {
		_, _ = cb.AnswerCallbackQuery("❌ Noto'g'ri so'rov", map[string]interface{}{"show_alert": true})
		return
	}

	card, err := models.GetCardByIDAndUser(int64(callbackData.ID), cb.AuthUser.ID)
	if err != nil {
		log.Printf("get card for delete confirm error: %v", err)
		_, _ = cb.AnswerCallbackQuery("❌ Xatolik yuz berdi", map[string]interface{}{"show_alert": true})
		return
	}
	if card == nil {
		_, _ = cb.AnswerCallbackQuery("❌ Karta topilmadi", map[string]interface{}{"show_alert": true})
		return
	}

	if err := models.DeleteCardByIDAndUser(card.ID, cb.AuthUser.ID); err != nil {
		log.Printf("delete card error: %v", err)
		_, _ = cb.AnswerCallbackQuery("❌ O'chirishda xatolik yuz berdi", map[string]interface{}{"show_alert": true})
		return
	}

	_, _ = cb.AnswerCallbackQuery("✅ Karta o'chirildi")
	_ = renderCardsMessage(cb, callbackData.Page)
}
