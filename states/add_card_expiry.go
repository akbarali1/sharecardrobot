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

func AddCardExpiryState(ctx *core.Update) {
	title := request.GetUserDataByChatId(ctx.Message.Chat.ID, "card_title")
	cardNumber := request.GetUserDataByChatId(ctx.Message.Chat.ID, "card_number")
	bankBinIDRaw := request.GetUserDataByChatId(ctx.Message.Chat.ID, "card_bank_bin_id")
	bankBinLabel := request.GetUserDataByChatId(ctx.Message.Chat.ID, "card_bank_bin_label")

	if title == "" || cardNumber == "" {
		ctx.ClearState()
		_, _ = ctx.SendMessage("❌ Karta ma'lumotlari topilmadi. /add_card ni qayta boshlang.", map[string]interface{}{
			"reply_markup": utils.MainKeyboard(),
		})
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

	bankBinID, err := strconv.ParseInt(strings.TrimSpace(bankBinIDRaw), 10, 64)
	if err != nil {
		bankBinID = 0
	}

	card, err := models.CreateCard(
		ctx.AuthUser.ID,
		title,
		cardNumber,
		utils.MaskCardNumber(cardNumber),
		expiryDate,
		utils.LastFour(cardNumber),
		bankBinID,
	)
	if err != nil {
		log.Printf("create card error: %v", err)
		_, _ = ctx.SendMessage("❌ Kartani saqlashda xatolik yuz berdi.", map[string]interface{}{
			"reply_markup": utils.MainKeyboard(),
		})
		return
	}

	expiryLine := ""
	if expiryDate != "" {
		expiryLine = fmt.Sprintf("\n<i>Amal qilish muddati:</i> <code>%s</code>", expiryDate)
	}

	ctx.ClearState()
	_, _ = ctx.SendMessage(fmt.Sprintf(
		"✅ Karta saqlandi.\n\n<b>%s</b>\n<code>%s</code>%s\n<i>%s</i>\n\nInline qidiruv uchun: <code>@%s %s</code>",
		html.EscapeString(card.Title),
		html.EscapeString(card.CardNumberMasked),
		expiryLine,
		html.EscapeString(bankBinLabel),
		core.BotUsername,
		html.EscapeString(card.LastFour),
	), map[string]interface{}{
		"reply_markup": utils.BuildSavedCardInlineMarkup(card.LastFour),
	})
	_, _ = ctx.SendMessage("Asosiy menyu", map[string]interface{}{
		"reply_markup": utils.MainKeyboard(),
	})
}
