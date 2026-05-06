# TODO

## Fix deterministic AES-GCM encryption for sensitive card fields

### Problem

`database/models/crypto.go` ichida `encryptSecretValue` AES-GCM uchun nonce'ni `HMAC(key, plaintext)` dan hosil qilmoqda. Bu encryption'ni deterministic qilib yuboradi, ya'ni bir xil plaintext har doim bir xil ciphertext beradi.

Natijada:
- bazada bir xil qiymatlar borligi bilinadi;
- PAN kabi sensitive field'lar uchun equality leakage paydo bo'ladi;
- `expiry_date` ham deterministic encrypt bo'lib qolgan, garchi unga exact-match lookup kerak bo'lmasa ham.

Bu immediate remote exploit emas, lekin DB leak yoki insider-read holatida sezilarli confidentiality weakness hisoblanadi.

### Recommended fix

1. `card_number` va `expiry_date` uchun random nonce bilan oddiy AES-GCM ishlatish.
2. Exact-match qidiruv kerak bo'lgan `card_number` uchun alohida fingerprint ustuni qo'shish.
3. Fingerprint sifatida `HMAC-SHA256(normalized PAN)` ishlatish va buning uchun encryption key'dan alohida secret key ishlatish.
4. SQL lookup'larni encrypted `card_number` bo'yicha emas, fingerprint bo'yicha qilish.
5. Mavjud ma'lumotlar uchun migration yozish:
   - eski yozuvlarni decrypt qilish;
   - yangi formatda qayta encrypt qilish;
   - fingerprint'ni to'ldirish;
   - query'larni yangi ustunga o'tkazish.

### Optional alternative

Deterministic lookup juda zarur bo'lsa, custom deterministic GCM o'rniga misuse-resistant mode ishlatish kerak:
- AES-SIV
- AES-GCM-SIV

Bu hozirgi custom nonce-derivation sxemasidan xavfsizroq.
