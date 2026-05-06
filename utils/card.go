package utils

import (
	"fmt"
	"html"
	"strings"
)

const SkipOptionalInputText = "O'tkazib yuborish"

func NormalizeCardNumber(raw string) string {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func IsSupportedCardNumber(number string) bool {
	number = NormalizeCardNumber(number)
	return len(number) == 16 || len(number) == 19
}

func IsValidLuhn(number string) bool {
	return luhn(NormalizeCardNumber(number))
}

func luhn(number string) bool {
	sum := 0
	for i, ch := range number {
		d := int(ch - '0')
		if (len(number)-i)%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return sum%10 == 0
}

func MaskCardNumber(number string) string {
	number = NormalizeCardNumber(number)
	if len(number) < 8 {
		return number
	}

	middleLength := len(number) - 8
	masked := number[:4] + strings.Repeat("*", middleLength) + number[len(number)-4:]
	return GroupCardNumber(masked)
}

func GroupCardNumber(number string) string {
	number = strings.TrimSpace(number)
	if number == "" {
		return ""
	}

	var parts []string
	for len(number) > 4 {
		parts = append(parts, number[:4])
		number = number[4:]
	}
	if number != "" {
		parts = append(parts, number)
	}
	return strings.Join(parts, " ")
}

func LastFour(number string) string {
	number = NormalizeCardNumber(number)
	if len(number) <= 4 {
		return number
	}
	return number[len(number)-4:]
}

func BuildCardMessage(title string, cardNumber string) string {
	return fmt.Sprintf("💳 <b>%s</b>\n<code>%s</code>",
		html.EscapeString(strings.TrimSpace(title)),
		html.EscapeString(GroupCardNumber(NormalizeCardNumber(cardNumber))),
	)
}

func BuildCardMessageWithExpiry(title string, cardNumber string, expiryDate string) string {
	message := BuildCardMessage(title, cardNumber)
	expiryDate = strings.TrimSpace(expiryDate)
	if expiryDate == "" {
		return message
	}

	return fmt.Sprintf("%s\n<i>Amal qilish muddati:</i> <code>%s</code>",
		message,
		html.EscapeString(expiryDate),
	)
}

func NormalizeCardExpiry(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}

	digits := b.String()
	if len(digits) != 4 {
		return ""
	}

	return digits[:2] + "/" + digits[2:]
}

func IsValidCardExpiry(raw string) bool {
	expiry := NormalizeCardExpiry(raw)
	if len(expiry) != 5 || expiry[2] != '/' {
		return false
	}

	month := (int(expiry[0]-'0') * 10) + int(expiry[1]-'0')
	return month >= 1 && month <= 12
}
