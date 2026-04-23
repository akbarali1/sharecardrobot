# share_card_robot

Telegram bot: https://t.me/share_card_robot

Bu bot foydalanuvchining karta raqamlarini saqlaydi va keyin ularni inline qidiruv orqali topib yuboradi.

## Minimal imkoniyatlar

- `/start` - user yaratish va yordam matni
- `/add_card` - yangi karta qo'shish
- `/my_cards` - kartalar ro'yxati
- inline qidiruv - `@share_card_robot 8600`
- `chosen_inline_result` statistikasi

## Sozlash

1. `.env.example` dan `.env` tayyorlang
2. MySQL bazani yarating
3. migratsiyalarni ishga tushiring:

```bash
mysql -u root -p share_card_robot < database/migrations/001_users.sql
mysql -u root -p share_card_robot < database/migrations/002_cards.sql
mysql -u root -p share_card_robot < database/migrations/003_inline_search_results.sql
mysql -u root -p share_card_robot < database/migrations/004_cards_add_bank_bin_id.sql
mysql -u root -p share_card_robot < database/migrations/005_cards_add_expiry_date.sql
mysql -u root -p share_card_robot < database/migrations/006_cards_encrypt_card_number.sql
```

4. karta raqamlarini encrypt qilish uchun AES-256 key yarating:

```bash
openssl rand -base64 32
```

Chiqqan qiymatni `.env` ichiga qo'shing:

```bash
CARD_ENCRYPTION_KEY=generated_base64_value
```

5. botni ishga tushiring:

```bash
go run .
```

`TELEGRAM_API_BASE` ixtiyoriy.
Agar local yoki self-hosted Telegram Bot API server ishlatsangiz, `.env` ichiga masalan:

```bash
TELEGRAM_API_BASE=http://127.0.0.1:8081
```

Agar bo'sh qoldirilsa, default `https://api.telegram.org` ishlatiladi.

## Hujjat

- minimal texnik topshiriq: [docs/task.md](./docs/task.md)
