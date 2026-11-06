# Telegram Serverless'ga ko'chirish

> **Holat:** 2026-yil 6-noyabrdan boshlab production bot **Telegram Serverless**'da ishlaydi.
> Bu repodagi Go kodi production'da ishlamaydi. U tarix va favqulodda rollback uchun saqlanadi.
>
> **Yangi kod:** [akbarali1/sharecardrobot-serverless](https://github.com/akbarali1/sharecardrobot-serverless) (private)

## Nima uchun

[Telegram Serverless](https://core.telegram.org/bots/serverless) bot kodini Telegram'ning o'z
infratuzilmasida (V8 isolate) ishga tushiradi. Bot API va bazasi platformaning ichida turadi.
Natijada quyidagilar endi kerak emas:

- VPS va systemd service;
- nginx orqali webhook;
- self-hosted Bot API server;
- MySQL.

## Nima o'zgardi

| | Eski (bu repo) | Yangi (Serverless) |
|---|---|---|
| Til | Go 1.25 | JavaScript (ES modules) |
| Hosting | VPS, `share_card_robot.service` | Telegram Serverless (`tgcloud`) |
| Bot API | Self-hosted local Bot API server (`TELEGRAM_API_BASE`) | Telegram cloud Bot API |
| Webhook | Go HTTP server, nginx va `TELEGRAM_WEBHOOK_SECRET` | Platformaning o'zi boshqaradi |
| Baza | MySQL | SQLite (har bir bot uchun alohida) |
| Dialog holati (state) | Xotiradagi map, restartda yo'qolardi | `user_states` jadvali, 24 soatlik TTL |
| Karta shifrlash | AES-256-GCM, deterministik nonce ([todo.md](../todo.md)) | AES-256-GCM, har safar tasodifiy nonce |
| Aniq raqam bo'yicha qidiruv | Shifrlangan qiymatni solishtirish | HMAC-SHA256 fingerprint (alohida kalit bilan) |
| Maxfiy kalitlar | `.env` | `tgcloud/lib/secrets.js` (gitignore'da, kod bilan birga deploy qilinadi) |
| Deploy | `go build` va `systemctl restart` | `npx tgcloud push` |

Foydalanuvchi uchun hech narsa o'zgarmagan:

- buyruqlar, matnlar, tugmalar va inline qidiruv aynan bir xil;
- router tartibi ham Go bilan bir xil, bu Go kodidan olingan "golden" testlar bilan tekshirilgan.

### `todo.md` dagi shifrlash muammosi

Serverless versiyada bu muammo hal qilingan:

- Har bir shifrlash yangi tasodifiy nonce ishlatadi. Bir xil karta raqami bazada har safar boshqacha ko'rinadi.
- Dublikat tekshiruvi va to'liq raqam bo'yicha qidiruv `card_number_fingerprint` ustuni orqali ishlaydi.
  Bu ustun `HMAC-SHA256(normalizatsiya qilingan raqam)` qiymati va alohida kalit bilan hisoblanadi.
- Kalitlar yangidan yaratilgan. Eski `CARD_ENCRYPTION_KEY` Telegram'ga yuborilmagan.

Bu repodagi Go kodida muammo tuzatilmagan, chunki u endi production'da ishlamaydi.

## Ma'lumotlar qanday ko'chirildi

| Jadval | Qatorlar | Izoh |
|---|---|---|
| `bank_bins` | 477 | Nazorat yig'indilari MySQL bilan bir xil |
| `users` | 10 | |
| `cards` | 17 | Qayta shifrlangan, 17 tasi ham platformada ochilishi tekshirilgan |
| `inline_search_results` | 16 | |
| `user_states` | 0 | Yangi jadval. Go'da state xotirada edi, shuning uchun ko'chiriladigan narsa yo'q |

MySQL bazasida boshqa jadval, view, procedure yoki trigger yo'q edi.

### MySQL va SQLite farqlari

- **Foreign key yo'q.** Platforma ruxsat bermaydi. `ON DELETE CASCADE` ilova kodida bajariladi:
  karta o'chirilganda uning `inline_search_results` qatorlari ham o'chiriladi.
- **Vaqtlar** `DATETIME` emas, unix sekund (`INTEGER`) ko'rinishida saqlanadi.
- **SQL o'zgarishlari:**
  - `ON DUPLICATE KEY UPDATE` o'rniga `onConflictDoUpdate`;
  - `NOW()` o'rniga `unixepoch()`;
  - BIN qidiruvi `substr(?, 1, length(bin_code)) = bin_code` ko'rinishida.
- **Karta nomi va bank bo'yicha qidiruv** JavaScript'da bajariladi. SQLite'ning `LIKE` operatori
  faqat ASCII uchun katta-kichik harfni farqlamaydi, MySQL'ning `utf8mb4_unicode_ci` esa kirill harflari
  uchun ham farqlamas edi.

### Vaqt zonasi haqida

MySQL server **UTC**da ishlagan. Go esa DSN'dagi `loc=Asia/Tashkent` tufayli vaqtlarni Toshkent vaqti
deb o'qigan, ya'ni ular 5 soatga siljigan edi. Eksport skripti `TIMESTAMPDIFF(SECOND, UTC_TIMESTAMP(), NOW())`
orqali haqiqiy offsetni aniqladi (natija `+00:00`). Shu tufayli Serverless bazasidagi vaqtlar to'g'ri.

## Ko'chirish bosqichlari

1. **Kod.** Bot JavaScript'da qayta yozildi va review qilindi. 98 ta test bor, ular orasida:
   - Go va JS shifrlash mosligi;
   - Go bot javoblari bilan parity.
2. **Platformani tekshirish.** Ikkalasi ham bot to'xtatilmasdan, deploy qilinmasdan bajarildi:
   - Runtime probe: `crypto.subtle` (AES-GCM, HMAC) ishlaydi. `atob`/`btoa` yo'q, lekin kod ularni ishlatmaydi.
   - DB probe: 13 ta tekshiruvning hammasi o'tdi.
3. **Sxema.** `tgcloud migrate --local` bilan 5 ta jadval va 8 ta indeks yaratildi.
4. **Zaxira.** MySQL'ning to'liq dump'i olindi (`mysqldump --single-transaction`).
5. **Sinov eksporti.** Bot ishlab turgan paytda ishga tushirildi. Kalitlar ishlashi va fayllarda ochiq
   karta raqami yo'qligi tekshirildi.
6. **O'tish.** Bot taxminan 1,5 daqiqa javob bermadi:
   - Go service to'xtatildi va `disable` qilindi.
   - Local Bot API serverda **faqat shu bot uchun** `deleteWebhook` va `logOut` qilindi. Bu server boshqa
     botlar uchun ham ishlaydi, shuning uchun uning o'zi to'xtatilmagan.
   - Yakuniy eksport qilindi. Bunda ma'lumot eski kalit bilan ochilib, yangi kalitlar bilan qayta shifrlandi.
   - Ma'lumotlar import qilindi, jadvallardagi qatorlar soni tekshirildi va `setMyCommands` bajarildi.
   - `npx tgcloud push` qilindi va webhook Telegram cloud'ga o'tdi.
7. **Tekshiruv:**
   - Webhook'da xato yo'q, sxema in sync.
   - 17 ta kartaning hammasi platformada ochiladi, fingerprint va oxirgi 4 raqam mos keladi.
   - `bank_bins` nazorat yig'indilari MySQL bilan bir xil.
8. **Tozalash.** Vaqtinchalik eksport fayllari va kalit fayllari serverdan va lokal kompyuterdan o'chirildi.

Bir martalik Go vositalari (eksport skripti va golden fixture generatori) yangi repoga `go-tools/`
papkasiga arxiv sifatida o'tkazildi.

## Bu repo endi nima uchun kerak

- Eski Go botning manba kodi va tarixi shu yerda.
- Rollback uchun kerak. Service serverda `disabled` holatda turibdi, MySQL bazasi va zaxirasi saqlangan.
- **Yangi o'zgarishlarni bu repoga emas**, [sharecardrobot-serverless](https://github.com/akbarali1/sharecardrobot-serverless) repoga kiriting.

## Rollback (favqulodda holat uchun)

> ⚠️ **Muhim:**
> - O'tishdan keyin Serverless'da yozilgan ma'lumotlar MySQL'da **yo'q**: yangi foydalanuvchilar, yangi va
>   tahrirlangan kartalar. Teskari eksport vositasi yo'q.
> - Yangi kartalar yangi kalit bilan shifrlangan, Go bot ularni o'qiy olmaydi.
>
> Rollback faqat jiddiy nosozlik bo'lganda qilinadi. Bu ma'lumotlar keyin qo'lda ko'chirilishi kerak.

1. Telegram cloud'dagi Serverless webhook'ini o'chiring va bot sessiyasini cloud'dan chiqaring:
   ```bash
   curl -sS "https://api.telegram.org/bot$BOT_TOKEN/deleteWebhook"
   curl -sS "https://api.telegram.org/bot$BOT_TOKEN/logOut"   # cloud'ga 10 daqiqa qayta kirib bo'lmaydi
   ```
   Agar Serverless webhook'ni qayta o'rnatib qo'ysa, BotFather → bot → Serverless'ni o'chiring.
   Bu qadam sinab ko'rilmagan.
2. Serverda Go botni yoqing. U ishga tushganda local Bot API serverda (`TELEGRAM_API_BASE`) webhook'ni
   o'zi o'rnatadi:
   ```bash
   sudo systemctl enable --now share_card_robot.service
   ```
3. Tekshiring:
   ```bash
   curl -sS "$TELEGRAM_API_BASE/bot$BOT_TOKEN/getWebhookInfo"   # url o'rnatilgan, pending 0, last_error yo'q
   ```

Serverless'ga qaytish uchun yangi repoda `npx tgcloud webhook sync` buyrug'ini ishga tushiring.
Undan oldin local Bot API'da shu bot uchun `deleteWebhook` va `logOut` qilinishi va Go service
to'xtatilishi kerak (yuqoridagi 6-bosqichga qarang).
