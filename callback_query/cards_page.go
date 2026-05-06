package callback_query

import (
	"log"

	"share_card_robot/core"
)

func CardsPageCallback(cb *core.CallbackQuery) {
	var callbackData core.CallbackData
	if err := jsonUnmarshalCallback(cb.Data, &callbackData); err != nil {
		_, _ = cb.AnswerCallbackQuery("❌ Noto'g'ri so'rov", map[string]interface{}{"show_alert": true})
		return
	}

	if callbackData.ID <= 0 {
		_, _ = cb.AnswerCallbackQuery("❌ Sahifa topilmadi", map[string]interface{}{"show_alert": true})
		return
	}

	_, _ = cb.AnswerCallbackQuery("")
	if err := renderCardsMessage(cb, callbackData.ID); err != nil {
		log.Printf("cards page callback error: %v", err)
	}
}
