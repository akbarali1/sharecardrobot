package commands

import (
	"fmt"
	"strings"

	"share_card_robot/core"
)

func HelpHandler(ctx *core.Update) {
	_, _ = ctx.SendMessage(buildHelpText())
}

func buildHelpText() string {
	botUsername := strings.TrimSpace(core.BotUsername)
	if botUsername == "" {
		botUsername = "share_card_robot"
	}

	return fmt.Sprintf(`💳 <b>Mening Kartalarim</b>

Bot sizning karta raqamlaringizni saqlaydi va keyin inline qidiruv orqali tez topib yuboradi.

<b>Komandalar:</b>
/start - botni ishga tushirish
/add_card - yangi karta qo'shish
/my_cards - saqlangan kartalar ro'yxati
/help - yordam

<b>Inline qidiruv:</b>
Istalgan chatda quyidagicha yozing:
<code>@%s uzcard</code>
<code>@%s humo</code>
<code>@%s b:kapital</code>
<code>@%s 3011</code>
<code>@%s asosiy</code>
<code>@%s exp 3011</code>
<code>@%s exp uzcard</code>

<b>Qidirish usullari:</b>
<code>uzcard</code> - faqat UZCARD kartalaringiz
<code>humo</code> - faqat HUMO kartalaringiz
<code>b:nomi</code> - bank nomi bo'yicha qidiruv
Raqam - karta raqami bo'yicha qidiruv
Oddiy matn - karta title bo'yicha qidiruv
<code>exp ...</code> - natijada amal qilish muddati ham ko'rsatiladi

<b>Misol:</b>
<code>@%s b:anor</code>
<code>@%s 3011</code>
<code>@%s exp 3011</code>
<code>@%s ish haqi</code>`, botUsername, botUsername, botUsername, botUsername, botUsername, botUsername, botUsername, botUsername, botUsername, botUsername, botUsername)
}
