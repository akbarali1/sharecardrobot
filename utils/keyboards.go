package utils

import (
	"fmt"
	"html"
	"strings"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/services"
)

func MainKeyboard() *core.ReplyKeyboardMarkup {
	return &core.ReplyKeyboardMarkup{
		Keyboard: [][]core.ReplyKeyboardButton{
			{
				{Text: "➕ Karta qo'shish"},
			},
			{
				{Text: "💳 Mening kartalarim"},
			},
		},
		ResizeKeyboard: true,
	}
}

func SkipOptionalKeyboard() *core.ReplyKeyboardMarkup {
	return &core.ReplyKeyboardMarkup{
		Keyboard: [][]core.ReplyKeyboardButton{
			{
				{Text: SkipOptionalInputText},
			},
		},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}

func BuildSavedCardInlineMarkup(query string) *core.InlineKeyboardMarkup {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}

	return &core.InlineKeyboardMarkup{
		InlineKeyboard: [][]core.KeyboardButton{
			{
				core.NewSwitchCurrentChatButton("🔍 Shu chatda qidirish", query),
			},
			{
				core.NewSwitchChatButton("📤 Boshqa chatga yuborish", query),
			},
		},
	}
}

func RenderCardsList(cards []*models.Card, page int, totalPages int) string {
	if len(cards) == 0 {
		return "💳 Sizda hali saqlangan kartalar yo'q.\n\nYangi karta qo'shish uchun /add_card yuboring."
	}

	lines := []string{"💳 <b>Mening kartalarim</b>", ""}
	for i, card := range cards {
		lines = append(lines, fmt.Sprintf("%d. <b>%s</b>", i+1, html.EscapeString(strings.TrimSpace(card.Title))))
		lines = append(lines, fmt.Sprintf("   <code>%s</code>", html.EscapeString(card.CardNumberMasked)))
		details := []string{card.CardTypeLabel(), card.BankNameLabel()}
		if details[0] != "" && details[1] != "" {
			lines = append(lines, fmt.Sprintf("   %s", html.EscapeString(strings.Join(details, " | "))))
		}
		lines = append(lines, "")
	}
	if totalPages > 1 {
		lines = append(lines, fmt.Sprintf("Sahifa: %d/%d", page, totalPages))
		lines = append(lines, "")
	}
	lines = append(lines, "Karta ustidagi tugma orqali taxrirlashingiz mumkin.")

	return strings.Join(lines, "\n")
}

func BuildCardsListMarkup(cards []*models.Card, page int, totalPages int) *core.InlineKeyboardMarkup {
	if len(cards) == 0 {
		return nil
	}

	buttons := make([]core.KeyboardButton, 0, len(cards))
	for i, card := range cards {
		buttons = append(buttons, core.KeyboardButton{
			Text: fmt.Sprintf("✏️ %d", i+1),
			CallbackData: core.CallbackData{
				Type: services.CallbackEditCardMenu,
				ID:   int(card.ID),
				Page: page,
			},
		})
	}

	markup := core.InlineKeyboard("", buttons, 4)
	if totalPages <= 1 {
		return markup
	}

	var navRow []core.KeyboardButton
	if page > 1 {
		navRow = append(navRow, core.NewCallbackButton("⬅️ Oldingi", services.CallbackCardsPage, page-1))
	}
	if page < totalPages {
		navRow = append(navRow, core.NewCallbackButton("Keyingi ➡️", services.CallbackCardsPage, page+1))
	}
	if len(navRow) > 0 {
		markup.InlineKeyboard = append(markup.InlineKeyboard, navRow)
	}

	return markup
}

func RenderDeletePrompt(card *models.Card) string {
	if card == nil {
		return "❌ Karta topilmadi."
	}

	return fmt.Sprintf(
		"⚠️ <b>Kartani o'chirasizmi?</b>\n\n<b>%s</b>\n<code>%s</code>",
		html.EscapeString(strings.TrimSpace(card.Title)),
		html.EscapeString(card.CardNumberMasked),
	)
}

func BuildDeletePromptMarkup(cardID int64, page int) *core.InlineKeyboardMarkup {
	return &core.InlineKeyboardMarkup{
		InlineKeyboard: [][]core.KeyboardButton{
			{
				core.NewCallbackButtonWithPage("✅ Ha, o'chirish", services.CallbackDeleteCardConfirm, int(cardID), page),
			},
			{
				core.NewCallbackButtonWithPage("⬅️ Bekor qilish", services.CallbackDeleteCardCancel, 0, page),
			},
		},
	}
}

func RenderEditCard(card *models.Card) string {
	if card == nil {
		return "❌ Karta topilmadi."
	}

	lines := []string{
		"✏️ <b>Kartani taxrirlash</b>",
		"",
		fmt.Sprintf("<b>Nomi:</b> %s", html.EscapeString(strings.TrimSpace(card.Title))),
		fmt.Sprintf("<b>Raqami:</b> <code>%s</code>", html.EscapeString(card.CardNumberMasked)),
	}

	expiry := card.ExpiryDateLabel()
	if expiry == "" {
		expiry = "Kiritilmagan"
	}
	lines = append(lines, fmt.Sprintf("<b>Amal qilish muddati:</b> <code>%s</code>", html.EscapeString(expiry)))

	details := []string{card.CardTypeLabel(), card.BankNameLabel()}
	if details[0] != "" || details[1] != "" {
		filtered := make([]string, 0, len(details))
		for _, detail := range details {
			detail = strings.TrimSpace(detail)
			if detail != "" {
				filtered = append(filtered, detail)
			}
		}
		if len(filtered) > 0 {
			lines = append(lines, fmt.Sprintf("<b>Qo'shimcha:</b> %s", html.EscapeString(strings.Join(filtered, " | "))))
		}
	}

	lines = append(lines, "", "Kerakli bo'limni tanlab taxrir qiling.")
	return strings.Join(lines, "\n")
}

func BuildEditCardMarkup(cardID int64, page int) *core.InlineKeyboardMarkup {
	return &core.InlineKeyboardMarkup{
		InlineKeyboard: [][]core.KeyboardButton{
			{
				core.NewCallbackButtonWithPage("✏️ Nomini taxrirlash", services.CallbackEditCardTitle, int(cardID), page),
			},
			{
				core.NewCallbackButtonWithPage("💳 Raqamini taxrirlash", services.CallbackEditCardNumber, int(cardID), page),
			},
			{
				core.NewCallbackButtonWithPage("📅 Muddatini taxrirlash", services.CallbackEditCardExpiry, int(cardID), page),
			},
			{
				core.NewCallbackButton("⬅️ Kartalar ro'yxati", services.CallbackCardsPage, page),
			},
		},
	}
}
