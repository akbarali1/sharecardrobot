package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"share_card_robot/database"
)

type Card struct {
	ID               int64
	UserID           int64
	Title            string
	CardNumber       string
	CardNumberMasked string
	ExpiryDate       sql.NullString
	LastFour         string
	BankBinID        int64
	BankName         sql.NullString
	CardTypeID       sql.NullInt64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func scanCard(rowScanner interface {
	Scan(dest ...any) error
}) (*Card, error) {
	var card Card
	err := rowScanner.Scan(
		&card.ID,
		&card.UserID,
		&card.Title,
		&card.CardNumber,
		&card.CardNumberMasked,
		&card.ExpiryDate,
		&card.LastFour,
		&card.BankBinID,
		&card.BankName,
		&card.CardTypeID,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func CreateCard(userID int64, title string, cardNumber string, masked string, expiryDate string, lastFour string, bankBinID int64) (*Card, error) {
	_, err := database.Exec(`INSERT INTO cards (user_id, title, card_number, card_number_masked, expiry_date, last_four, bank_bin_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		userID,
		strings.TrimSpace(title),
		strings.TrimSpace(cardNumber),
		strings.TrimSpace(masked),
		nullIfEmpty(expiryDate),
		strings.TrimSpace(lastFour),
		bankBinID,
	)
	if err != nil {
		return nil, err
	}

	row := database.QueryRow(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ? AND c.card_number = ? LIMIT 1`, userID, strings.TrimSpace(cardNumber))
	return scanCard(row)
}

func CardExistsByUserAndNumber(userID int64, cardNumber string) (bool, error) {
	row := database.QueryRow(`SELECT id FROM cards WHERE user_id = ? AND card_number = ? LIMIT 1`, userID, strings.TrimSpace(cardNumber))
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return id > 0, nil
}

func ListCardsByUser(userID int64, limit int) ([]*Card, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ? ORDER BY c.updated_at DESC, c.id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func CountCardsByUser(userID int64) (int, error) {
	row := database.QueryRow(`SELECT COUNT(*) FROM cards WHERE user_id = ?`, userID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func ListCardsByUserPage(userID int64, limit int, offset int) ([]*Card, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ? ORDER BY c.updated_at DESC, c.id DESC LIMIT ? OFFSET ?`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func SearchCardsByUser(userID int64, query string, limit int) ([]*Card, error) {
	if limit <= 0 {
		limit = 10
	}

	query = strings.TrimSpace(query)
	if query == "" {
		return ListCardsByUserByChoicePriority(userID, limit)
	}
	if cardTypeID := inlineQueryCardTypeID(query); cardTypeID > 0 {
		return listCardsByUserAndCardType(userID, cardTypeID, limit)
	}
	if bankNameQuery, ok := inlineQueryBankName(query); ok {
		return listCardsByUserAndBankName(userID, bankNameQuery, limit)
	}
	if numberQuery, ok := inlineQueryCardNumber(query); ok {
		return listCardsByUserAndCardNumber(userID, numberQuery, limit)
	}

	titleLike := "%" + query + "%"
	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ?
		  AND c.title LIKE ?
		ORDER BY
		  CASE
		    WHEN c.title LIKE ? THEN 0
		    ELSE 1
		  END,
		  c.updated_at DESC,
		  c.id DESC
		LIMIT ?`,
		userID,
		titleLike,
		query+"%",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func ListCardsByUserByChoicePriority(userID int64, limit int) ([]*Card, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		LEFT JOIN (
			SELECT user_id, card_id, MAX(chosen_count) AS chosen_count, MAX(shown_count) AS shown_count
			FROM inline_search_results
			WHERE user_id = ?
			GROUP BY user_id, card_id
		) stats ON stats.user_id = c.user_id AND stats.card_id = c.id
		WHERE c.user_id = ?
		ORDER BY
		  CASE
		    WHEN COALESCE(stats.chosen_count, 0) > 0 THEN 0
		    ELSE 1
		  END,
		  COALESCE(stats.chosen_count, 0) DESC,
		  COALESCE(stats.shown_count, 0) DESC,
		  c.updated_at DESC,
		  c.id DESC
		LIMIT ?`, userID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func listCardsByUserAndCardNumber(userID int64, numberQuery string, limit int) ([]*Card, error) {
	numberQuery = strings.TrimSpace(numberQuery)
	if numberQuery == "" {
		return []*Card{}, nil
	}

	numberLike := "%" + numberQuery + "%"
	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ?
		  AND c.card_number LIKE ?
		ORDER BY
		  CASE
		    WHEN c.card_number LIKE ? THEN 0
		    ELSE 1
		  END,
		  c.updated_at DESC,
		  c.id DESC
		LIMIT ?`, userID, numberLike, numberQuery+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func listCardsByUserAndCardType(userID int64, cardTypeID int64, limit int) ([]*Card, error) {
	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ? AND bb.card_type_id = ?
		ORDER BY c.updated_at DESC, c.id DESC
		LIMIT ?`, userID, cardTypeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func listCardsByUserAndBankName(userID int64, bankName string, limit int) ([]*Card, error) {
	bankName = strings.TrimSpace(bankName)
	if bankName == "" {
		return []*Card{}, nil
	}

	bankNameLike := "%" + bankName + "%"
	rows, err := database.Query(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.user_id = ? AND bb.name LIKE ?
		ORDER BY
		  CASE
		    WHEN bb.name LIKE ? THEN 0
		    ELSE 1
		  END,
		  c.updated_at DESC,
		  c.id DESC
		LIMIT ?`, userID, bankNameLike, bankName+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, rows.Err()
}

func GetCardByIDAndUser(cardID int64, userID int64) (*Card, error) {
	row := database.QueryRow(`SELECT c.id, c.user_id, c.title, c.card_number, c.card_number_masked, c.expiry_date, c.last_four, c.bank_bin_id, bb.name, bb.card_type_id, c.created_at, c.updated_at
		FROM cards c
		LEFT JOIN bank_bins bb ON bb.id = c.bank_bin_id
		WHERE c.id = ? AND c.user_id = ? LIMIT 1`, cardID, userID)
	card, err := scanCard(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return card, nil
}

func DeleteCardByIDAndUser(cardID int64, userID int64) error {
	result, err := database.Exec(`DELETE FROM cards WHERE id = ? AND user_id = ? LIMIT 1`, cardID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}
	return nil
}

func digitsOnly(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func inlineQueryCardTypeID(query string) int64 {
	normalized := strings.NewReplacer(" ", "", "-", "", "_", "").Replace(strings.ToLower(strings.TrimSpace(query)))

	switch normalized {
	case "uzcard":
		return 1
	case "humo":
		return 2
	default:
		return 0
	}
}

func inlineQueryBankName(query string) (string, bool) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return "", false
	}
	if !strings.HasPrefix(strings.ToLower(query), "b:") {
		return "", false
	}

	return strings.TrimSpace(query[2:]), true
}

func inlineQueryCardNumber(query string) (string, bool) {
	query = strings.TrimSpace(query)
	if query == "" {
		return "", false
	}

	digitsQuery := digitsOnly(query)
	if digitsQuery == "" || digitsQuery != query {
		return "", false
	}

	return digitsQuery, true
}

func (c *Card) CardTypeLabel() string {
	if c == nil || !c.CardTypeID.Valid {
		return "Noma'lum"
	}

	switch c.CardTypeID.Int64 {
	case 1:
		return "UZCARD"
	case 2:
		return "HUMO"
	default:
		return "Noma'lum"
	}
}

func (c *Card) BankNameLabel() string {
	if c == nil || !c.BankName.Valid || strings.TrimSpace(c.BankName.String) == "" {
		return "Bank topilmadi"
	}

	return strings.TrimSpace(c.BankName.String)
}

func (c *Card) BankBins() *BankBin {
	if c == nil || c.BankBinID <= 0 {
		return nil
	}

	row := database.QueryRow(`SELECT id, name, card_name, bin_code FROM bank_bins WHERE id = ? LIMIT 1`, c.BankBinID)
	bankBin, err := scanBankBin(row)
	if err != nil {
		//if errors.Is(err, sql.ErrNoRows) {
		//	return nil
		//}
		return nil
	}

	return bankBin
}

func (c *Card) ExpiryDateLabel() string {
	if c == nil || !c.ExpiryDate.Valid || strings.TrimSpace(c.ExpiryDate.String) == "" {
		return ""
	}

	return strings.TrimSpace(c.ExpiryDate.String)
}
