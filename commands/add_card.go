package commands

import "share_card_robot/core"

func AddCardHandler(ctx *core.Update) {
	if ctx == nil {
		return
	}

	ctx.ClearState()
	ctx.SetState("add_card_name")
	_, _ = ctx.SendMessage("📝 Karta nomini yuboring.\n\nMasalan: <code>Asosiy karta</code>", map[string]interface{}{
		"reply_markup": &core.ReplyKeyboardRemove{RemoveKeyboard: true},
	})
}
