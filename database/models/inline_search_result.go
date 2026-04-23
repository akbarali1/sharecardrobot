package models

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"share_card_robot/database"
)

type InlineSearchResult struct {
	ID           int64
	ResultID     string
	UserID       int64
	CardID       int64
	Title        string
	Description  string
	MessageText  string
	ShownCount   int
	ChosenCount  int
	LastQuery    sql.NullString
	LastChosenAt sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func scanInlineSearchResult(rowScanner interface {
	Scan(dest ...any) error
}) (*InlineSearchResult, error) {
	var item InlineSearchResult
	err := rowScanner.Scan(
		&item.ID,
		&item.ResultID,
		&item.UserID,
		&item.CardID,
		&item.Title,
		&item.Description,
		&item.MessageText,
		&item.ShownCount,
		&item.ChosenCount,
		&item.LastQuery,
		&item.LastChosenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func UpsertInlineSearchResult(resultID string, userID int64, cardID int64, title string, description string, messageText string, query string) error {
	_, err := database.Exec(`INSERT INTO inline_search_results
		(result_id, user_id, card_id, title, description, message_text, shown_count, last_query, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			card_id = VALUES(card_id),
			title = VALUES(title),
			description = VALUES(description),
			message_text = VALUES(message_text),
			shown_count = shown_count + 1,
			last_query = VALUES(last_query),
			updated_at = NOW()`,
		strings.TrimSpace(resultID),
		userID,
		cardID,
		strings.TrimSpace(title),
		strings.TrimSpace(description),
		strings.TrimSpace(messageText),
		nullIfEmpty(query),
	)
	return err
}

func IncrementInlineSearchResultChoice(resultID string, query string, userID int64) error {
	normalizedResultID := normalizeInlineSearchResultID(resultID)

	result, err := database.Exec(`UPDATE inline_search_results
		SET chosen_count = chosen_count + 1,
			last_query = ?,
			last_chosen_at = NOW(),
			updated_at = NOW()
		WHERE result_id = ? AND user_id = ?`,
		nullIfEmpty(query),
		normalizedResultID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		return nil
	}

	cardID, err := parseInlineSearchCardID(resultID)
	if err != nil {
		return err
	}

	card, err := GetCardByIDAndUser(cardID, userID)
	if err != nil {
		return err
	}
	if card == nil {
		return fmt.Errorf("card not found for result: %s", resultID)
	}

	if err := UpsertInlineSearchResult(
		normalizedResultID,
		userID,
		card.ID,
		card.Title,
		card.CardNumberMasked,
		fmt.Sprintf("💳 <b>%s</b>\n<code>%s</code>",
			html.EscapeString(strings.TrimSpace(card.Title)),
			html.EscapeString(strings.TrimSpace(card.CardNumber)),
		),
		query,
	); err != nil {
		return err
	}

	_, err = database.Exec(`UPDATE inline_search_results
		SET chosen_count = chosen_count + 1,
			last_query = ?,
			last_chosen_at = NOW(),
			updated_at = NOW()
		WHERE result_id = ? AND user_id = ?`,
		nullIfEmpty(query),
		normalizedResultID,
		userID,
	)
	return err
}

func TopInlineSearchResultsByUser(userID int64, limit int) ([]*InlineSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := database.Query(`SELECT id, result_id, user_id, card_id, title, description, message_text,
			shown_count, chosen_count, last_query, last_chosen_at, created_at, updated_at
		FROM inline_search_results
		WHERE user_id = ?
		ORDER BY chosen_count DESC, shown_count DESC, updated_at DESC
		LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*InlineSearchResult
	for rows.Next() {
		item, err := scanInlineSearchResult(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func parseInlineSearchCardID(resultID string) (int64, error) {
	parts := strings.Split(normalizeInlineSearchResultID(resultID), "_")
	if len(parts) != 2 || parts[0] != "card" {
		return 0, errors.New("invalid inline search result id")
	}

	cardID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || cardID <= 0 {
		return 0, errors.New("invalid inline search card id")
	}
	return cardID, nil
}

func normalizeInlineSearchResultID(resultID string) string {
	resultID = strings.TrimSpace(resultID)
	return strings.TrimSuffix(resultID, "_exp")
}
