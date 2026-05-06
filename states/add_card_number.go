package states

import (
	"fmt"
	"log"
	"strconv"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/services/request"
	"share_card_robot/utils"
)

func AddCardNumberState(ctx *core.Update) {
	title := request.GetUserDataByChatId(ctx.Message.Chat.ID, "card_title")
	if title == "" {
		ctx.ClearState()
		_, _ = ctx.SendMessage("❌ Karta nomi topilmadi. /add_card ni qayta boshlang.")
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

	exists, err := models.CardExistsByUserAndNumber(ctx.AuthUser.ID, cardNumber)
	if err != nil {
		log.Printf("card exists check error: %v", err)
		_, _ = ctx.SendMessage("❌ Kartani tekshirishda xatolik yuz berdi.")
		return
	}
	if exists {
		_, _ = ctx.SendMessage("ℹ️ Bu karta raqamini kiritgansiz. Nomini o'zgartirmoqchi bo'lsangiz, avval o'chiring.")
		return
	}

	bankBin, err := models.FindBankBinByCardNumber(cardNumber)
	if err != nil {
		log.Printf("bank bin lookup error: %v", err)
		_, _ = ctx.SendMessage("❌ Karta BIN ma'lumotini topishda xatolik yuz berdi.")
		return
	}

	bankBinID := int64(0)
	bankBinLabel := "BIN topilmadi"
	if bankBin != nil {
		bankBinID = bankBin.ID
		bankBinLabel = bankBin.DisplayName()
	}

	request.SetUserDataByChatId(ctx.Message.Chat.ID, "card_number", cardNumber)
	request.SetUserDataByChatId(ctx.Message.Chat.ID, "card_bank_bin_id", strconv.FormatInt(bankBinID, 10))
	request.SetUserDataByChatId(ctx.Message.Chat.ID, "card_bank_bin_label", bankBinLabel)
	ctx.SetState("add_card_expiry")
	_, _ = ctx.SendMessage(
		fmt.Sprintf("📅 Amal qilish muddatini kiriting.\n\nMasalan: <code>12/27</code>\nKiritishni xohlamasangiz, <b>%s</b> tugmasini bosing.", utils.SkipOptionalInputText),
		map[string]interface{}{"reply_markup": utils.SkipOptionalKeyboard()},
	)
}
