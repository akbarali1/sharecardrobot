package models

import (
	"database/sql"
	"errors"
	"strings"

	"share_card_robot/database"
)

type BankBin struct {
	ID       int64
	Name     string
	CardName string
	BinCode  string
}

func scanBankBin(rowScanner interface {
	Scan(dest ...any) error
}) (*BankBin, error) {
	var bankBin BankBin
	err := rowScanner.Scan(
		&bankBin.ID,
		&bankBin.Name,
		&bankBin.CardName,
		&bankBin.BinCode,
	)
	if err != nil {
		return nil, err
	}
	return &bankBin, nil
}

func FindBankBinByCardNumber(cardNumber string) (*BankBin, error) {
	cardNumber = digitsOnly(cardNumber)
	if cardNumber == "" {
		return nil, nil
	}

	row := database.QueryRow(`SELECT id, name, card_name, bin_code
		FROM bank_bins
		WHERE CAST(LEFT(?, CHAR_LENGTH(bin_code)) AS BINARY) = CAST(bin_code AS BINARY)
		ORDER BY CHAR_LENGTH(bin_code) DESC, id ASC
		LIMIT 1`, cardNumber)

	bankBin, err := scanBankBin(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return bankBin, nil
}

func (b *BankBin) DisplayName() string {
	if b == nil {
		return ""
	}

	return strings.TrimSpace(strings.TrimSpace(b.Name) + " " + strings.TrimSpace(b.CardName))
}

func (b *BankBin) DisplayOnlyName() string {
	if b == nil {
		return ""
	}

	return strings.TrimSpace(strings.TrimSpace(b.Name))
}
