package states

import (
	"fmt"
	"html"
	"log"
	"strconv"
	"strings"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/services/request"
	"share_card_robot/utils"
)

type editCardContext struct {
	Card      *models.Card
	Page      int
	MessageID int
}

func EditCardTitleState(ctx *core.Update) {
	editCtx, ok := loadEditCardContext(ctx)
	if !ok {
		return
	}

	title := strings.TrimSpace(ctx.Message.Text)
	if title == "" {
		_, _ = ctx.SendMessage("❌ Karta nomi bo'sh bo'lmasligi kerak. Qayta yuboring.")
		return
	}

	card, err := models.UpdateCardTitleByIDAndUser(editCtx.Card.ID, ctx.AuthUser.ID, title)
	if err != nil {
		log.Printf("update card title error: %v", err)
		_, _ = ctx.SendMessage("❌ Karta nomini saqlashda xatolik yuz berdi.")
		return
	}

	finishCardEdit(ctx, editCtx, card, "✅ Karta nomi yangilandi.")
}

func EditCardNumberState(ctx *core.Update) {
	editCtx, ok := loadEditCardContext(ctx)
	if !ok {
		return
	}

	cardNumber := utils.NormalizeCardNumber(ctx.Message.Text)
	if !utils.IsSupportedCardNumber(cardNumber) {
		_, _ = ctx.SendMessage("❌ Karta raqami noto'g'ri. 16 yoki 19 xonali raqam yuboring.")
		return
	}

	if !utils.IsValidLuhn(cardNumber) {
		_, _ = ctx.SendMessage("❌ Karta raqami xato kiritdingiz.")
		return
	}

	exists, err := models.CardExistsByUserAndNumberExceptID(ctx.AuthUser.ID, editCtx.Card.ID, cardNumber)
	if err != nil {
		log.Printf("card exists check error: %v", err)
		_, _ = ctx.SendMessage("❌ Kartani tekshirishda xatolik yuz berdi.")
		return
	}
	if exists {
		_, _ = ctx.SendMessage("ℹ️ Bu karta raqami sizda boshqa karta sifatida saqlangan.")
		return
	}

	bankBin, err := models.FindBankBinByCardNumber(cardNumber)
	if err != nil {
		log.Printf("bank bin lookup error: %v", err)
		_, _ = ctx.SendMessage("❌ Karta BIN ma'lumotini topishda xatolik yuz berdi.")
		return
	}

	bankBinID := int64(0)
	if bankBin != nil {
		bankBinID = bankBin.ID
	}

	card, err := models.UpdateCardNumberByIDAndUser(
		editCtx.Card.ID,
		ctx.AuthUser.ID,
		cardNumber,
		utils.MaskCardNumber(cardNumber),
		utils.LastFour(cardNumber),
		bankBinID,
	)
	if err != nil {
		log.Printf("update card number error: %v", err)
		_, _ = ctx.SendMessage("❌ Karta raqamini saqlashda xatolik yuz berdi.")
		return
	}

	finishCardEdit(ctx, editCtx, card, "✅ Karta raqami yangilandi.")
}

func EditCardExpiryState(ctx *core.Update) {
	editCtx, ok := loadEditCardContext(ctx)
	if !ok {
		return
	}

	expiryInput := strings.TrimSpace(ctx.Message.Text)
	expiryDate := ""
	if expiryInput != utils.SkipOptionalInputText {
		if !utils.IsValidCardExpiry(expiryInput) {
			_, _ = ctx.SendMessage("❌ Amal qilish muddati noto'g'ri. <code>MM/YY</code> yoki <code>MMYY</code> formatida yuboring.", map[string]interface{}{
				"reply_markup": utils.SkipOptionalKeyboard(),
			})
			return
		}
		expiryDate = utils.NormalizeCardExpiry(expiryInput)
	}

	card, err := models.UpdateCardExpiryByIDAndUser(editCtx.Card.ID, ctx.AuthUser.ID, expiryDate)
	if err != nil {
		log.Printf("update card expiry error: %v", err)
		_, _ = ctx.SendMessage("❌ Amal qilish muddatini saqlashda xatolik yuz berdi.")
		return
	}

	finishCardEdit(ctx, editCtx, card, "✅ Amal qilish muddati yangilandi.")
}

func loadEditCardContext(ctx *core.Update) (*editCardContext, bool) {
	cardIDRaw := request.GetUserDataByChatId(ctx.Message.Chat.ID, "edit_card_id")
	pageRaw := request.GetUserDataByChatId(ctx.Message.Chat.ID, "edit_cards_page")
	messageIDRaw := request.GetUserDataByChatId(ctx.Message.Chat.ID, "edit_message_id")

	cardID, err := strconv.ParseInt(strings.TrimSpace(cardIDRaw), 10, 64)
	if err != nil || cardID <= 0 {
		ctx.ClearState()
		_, _ = ctx.SendMessage("❌ Tahrirlanayotgan karta topilmadi. /my_cards orqali qayta tanlang.", map[string]interface{}{
			"reply_markup": utils.MainKeyboard(),
		})
		return nil, false
	}

	page, _ := strconv.Atoi(strings.TrimSpace(pageRaw))
	if page <= 0 {
		page = 1
	}

	messageID, _ := strconv.Atoi(strings.TrimSpace(messageIDRaw))

	card, err := models.GetCardByIDAndUser(cardID, ctx.AuthUser.ID)
	if err != nil {
		log.Printf("get card for state edit error: %v", err)
		_, _ = ctx.SendMessage("❌ Kartani olishda xatolik yuz berdi.")
		return nil, false
	}
	if card == nil {
		ctx.ClearState()
		_, _ = ctx.SendMessage("❌ Karta topilmadi. /my_cards orqali qayta tanlang.", map[string]interface{}{
			"reply_markup": utils.MainKeyboard(),
		})
		return nil, false
	}

	return &editCardContext{
		Card:      card,
		Page:      page,
		MessageID: messageID,
	}, true
}

func finishCardEdit(ctx *core.Update, editCtx *editCardContext, card *models.Card, successText string) {
	ctx.ClearState()

	if card != nil && editCtx.MessageID > 0 {
		if _, err := core.EditMessageText(ctx.Message.Chat.ID, editCtx.MessageID, utils.RenderEditCard(card), map[string]interface{}{
			"reply_markup": utils.BuildEditCardMarkup(card.ID, editCtx.Page),
		}); err != nil {
			log.Printf("edit card message refresh error: %v", err)
		}
	}

	if card == nil {
		_, _ = ctx.SendMessage("✅ O'zgarish saqlandi, lekin kartani qayta yuklab bo'lmadi. /my_cards orqali yangilangan holatini ko'ring.", map[string]interface{}{
			"reply_markup": utils.MainKeyboard(),
		})
		return
	}

	_, _ = ctx.SendMessage(fmt.Sprintf(
		"%s\n\n<b>%s</b>\n<code>%s</code>\n<i>Amal qilish muddati:</i> <code>%s</code>",
		successText,
		html.EscapeString(card.Title),
		html.EscapeString(card.CardNumberMasked),
		html.EscapeString(renderExpiryLabel(card)),
	))
}

func renderExpiryLabel(card *models.Card) string {
	if card == nil {
		return "Kiritilmagan"
	}

	expiry := card.ExpiryDateLabel()
	if expiry == "" {
		return "Kiritilmagan"
	}
	return expiry
}
