package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"share_card_robot/database"
)

type User struct {
	ID             int64
	TelegramUserID int64
	Username       sql.NullString
	FirstName      string
	LastName       sql.NullString
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) GetUsername() string {
	if u.Username.Valid {
		return u.Username.String
	}
	return ""
}

func (u *User) GetLastName() string {
	if u.LastName.Valid {
		return u.LastName.String
	}
	return ""
}

func (u *User) DisplayName() string {
	fullName := strings.TrimSpace(u.FirstName + " " + u.GetLastName())
	if fullName != "" {
		return fullName
	}
	if username := u.GetUsername(); username != "" {
		return "@" + username
	}
	return "Foydalanuvchi"
}

func scanUser(rowScanner interface {
	Scan(dest ...any) error
}) (*User, error) {
	var user User
	err := rowScanner.Scan(
		&user.ID,
		&user.TelegramUserID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByTelegramID(telegramUserID int64) (*User, error) {
	row := database.QueryRow(`SELECT id, telegram_user_id, username, first_name, last_name, created_at, updated_at
		FROM users WHERE telegram_user_id = ? LIMIT 1`, telegramUserID)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func GetOrCreateTelegramUser(telegramUserID int64, username string, firstName string, lastName string) (*User, error) {
	user, err := GetUserByTelegramID(telegramUserID)
	if err != nil {
		return nil, err
	}

	username = strings.TrimSpace(strings.TrimPrefix(username, "@"))
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	if user == nil {
		_, err := database.Exec(`INSERT INTO users (telegram_user_id, username, first_name, last_name, created_at, updated_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())`,
			telegramUserID,
			nullIfEmpty(username),
			firstName,
			nullIfEmpty(lastName),
		)
		if err != nil {
			return nil, err
		}
		return GetUserByTelegramID(telegramUserID)
	}

	if user.GetUsername() != username || user.FirstName != firstName || user.GetLastName() != lastName {
		_, err = database.Exec(`UPDATE users
			SET username = ?, first_name = ?, last_name = ?, updated_at = NOW()
			WHERE id = ?`,
			nullIfEmpty(username),
			firstName,
			nullIfEmpty(lastName),
			user.ID,
		)
		if err != nil {
			return nil, err
		}
		return GetUserByTelegramID(telegramUserID)
	}

	return user, nil
}

func nullIfEmpty(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}
