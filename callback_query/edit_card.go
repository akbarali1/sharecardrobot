package callback_query

import (
	"html"
	"log"
	"strconv"
	"strings"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/services/request"
	"share_card_robot/utils"
)

func EditCardMenuCallback(cb *core.CallbackQuery) {
	var callbackData core.CallbackData
	if err := jsonUnmarshalCallback(cb.Data, &callbackData); err != nil {
		_, _ = cb.AnswerCallbackQuery("❌ Noto'g'ri so'rov", map[string]interface{}{"show_alert": true})
		return
	}

	card, err := models.GetCardByIDAndUser(int64(callbackData.ID), cb.AuthUser.ID)
	if err != nil {
		log.Printf("get card for edit menu error: %v", err)
		_, _ = cb.AnswerCallbackQuery("❌ Xatolik yuz berdi", map[string]interface{}{"show_alert": true})
		return
	}
	if card == nil {
		_, _ = cb.AnswerCallbackQuery("❌ Karta topilmadi", map[string]interface{}{"show_alert": true})
		return
	}

	_, _ = cb.AnswerCallbackQuery("✏️ Taxrirlash menyusi")
	_, _ = cb.BaseEditMessageText(cb.Message.MessageID, utils.RenderEditCard(card), map[string]interface{}{
		"reply_markup": utils.BuildEditCardMarkup(card.ID, callbackData.Page),
	})
}

func EditCardTitleCallback(cb *core.CallbackQuery) {
	startCardEditFlow(cb, "edit_card_title", "📝 Yangi karta nomini yuboring.", func(card *models.Card) string {
		return "Joriy nomi: <code>" + html.EscapeString(strings.TrimSpace(card.Title)) + "</code>"
	})
}

func EditCardNumberCallback(cb *core.CallbackQuery) {
	startCardEditFlow(cb, "edit_card_number", "💳 Yangi karta raqamini yuboring.", func(card *models.Card) string {
		return "Joriy raqami: <code>" + html.EscapeString(card.CardNumberMasked) + "</code>"
	})
}

func EditCardExpiryCallback(cb *core.CallbackQuery) {
	startCardEditFlow(cb, "edit_card_expiry", "📅 Yangi amal qilish muddatini yuboring.\n\nMasalan: <code>12/27</code>\nTozalash uchun <b>"+utils.SkipOptionalInputText+"</b> tugmasini bosing.", func(card *models.Card) string {
		expiry := card.ExpiryDateLabel()
		if expiry == "" {
			expiry = "kiritilmagan"
		}
		return "Joriy muddati: <code>" + html.EscapeString(expiry) + "</code>"
	}, map[string]interface{}{"reply_markup": utils.SkipOptionalKeyboard()})
}

func startCardEditFlow(cb *core.CallbackQuery, state string, prompt string, currentValue func(*models.Card) string, opts ...map[string]interface{}) {
	var callbackData core.CallbackData
	if err := jsonUnmarshalCallback(cb.Data, &callbackData); err != nil {
		_, _ = cb.AnswerCallbackQuery("❌ Noto'g'ri so'rov", map[string]interface{}{"show_alert": true})
		return
	}

	card, err := models.GetCardByIDAndUser(int64(callbackData.ID), cb.AuthUser.ID)
	if err != nil {
		log.Printf("get card for edit flow error: %v", err)
		_, _ = cb.AnswerCallbackQuery("❌ Xatolik yuz berdi", map[string]interface{}{"show_alert": true})
		return
	}
	if card == nil {
		_, _ = cb.AnswerCallbackQuery("❌ Karta topilmadi", map[string]interface{}{"show_alert": true})
		return
	}

	request.SetUserDataByChatId(cb.Message.Chat.ID, "edit_card_id", strconv.FormatInt(card.ID, 10))
	request.SetUserDataByChatId(cb.Message.Chat.ID, "edit_cards_page", strconv.Itoa(callbackData.Page))
	request.SetUserDataByChatId(cb.Message.Chat.ID, "edit_message_id", strconv.Itoa(cb.Message.MessageID))
	request.SetUserStateByChatId(cb.Message.Chat.ID, state)

	_, _ = cb.AnswerCallbackQuery("✏️ Qiymatni yuboring")
	message := prompt
	if currentValue != nil {
		message += "\n\n" + currentValue(card)
	}
	if len(opts) > 0 {
		_, _ = cb.SendMessage(message, opts[0])
		return
	}
	_, _ = cb.SendMessage(message)
}
