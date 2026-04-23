package commands

import (
	"fmt"
	"log"
	"strings"

	"share_card_robot/core"
	"share_card_robot/database/models"
	"share_card_robot/utils"
)

func InlineSearchHandler(ctx *core.InlineQuery) {
	rawQuery := strings.TrimSpace(ctx.Query)
	query, showExpiry := parseInlineQuery(rawQuery)

	cards, err := models.SearchCardsByUser(ctx.AuthUser.ID, query, 10)
	if err != nil {
		log.Printf("inline search cards error: %v", err)
		answerEmptyInline(ctx.ID, "❌ Qidiruvda xatolik yuz berdi")
		return
	}

	if len(cards) == 0 {
		if query == "" {
			answerEmptyInline(ctx.ID, "💳 Avval kartalarni botga saqlang")
			return
		}
		answerEmptyInline(ctx.ID, "❌ Hech narsa topilmadi")
		return
	}

	articles := make([]core.InlineQueryResultArticle, 0, len(cards))
	for _, card := range cards {
		article := buildCardArticle(card, showExpiry)
		storedArticle := buildCardArticle(card, false)
		if err := models.UpsertInlineSearchResult(
			storedArticle.ID,
			ctx.AuthUser.ID,
			card.ID,
			storedArticle.Title,
			storedArticle.Description,
			storedArticle.InputMessageContent.MessageText,
			rawQuery,
		); err != nil {
			log.Printf("inline result upsert error: %v", err)
		}
		articles = append(articles, article)
	}

	if _, err := core.AnswerInlineQuery(ctx.ID, articles, map[string]interface{}{
		"cache_time":  0,
		"is_personal": true,
	}); err != nil {
		log.Printf("answer inline query error: %v", err)
	}
}

func answerEmptyInline(inlineID string, message string) {
	opts := map[string]interface{}{
		"cache_time":  0,
		"is_personal": true,
	}
	if strings.TrimSpace(message) != "" {
		opts["switch_pm_text"] = message
		opts["switch_pm_parameter"] = "start"
	}

	if _, err := core.AnswerInlineQuery(inlineID, []core.InlineQueryResultArticle{}, opts); err != nil {
		log.Printf("answer empty inline query error: %v", err)
	}
}

func buildCardArticle(card *models.Card, showExpiry bool) core.InlineQueryResultArticle {
	messageText := utils.BuildCardMessage(card.Title, card.CardNumber)
	description := card.CardNumberMasked
	if showExpiry && card.ExpiryDateLabel() != "" {
		messageText = utils.BuildCardMessageWithExpiry(card.Title, card.CardNumber, card.ExpiryDateLabel())
		description = fmt.Sprintf("%s | %s", card.CardNumberMasked, card.ExpiryDateLabel())
	}

	bins := card.BankBins()
	if bins != nil && bins.DisplayName() != "" {
		description = fmt.Sprintf("%s\n%s", description, bins.DisplayOnlyName())
	}

	return core.InlineQueryResultArticle{
		Type:        "article",
		ID:          inlineResultID(card.ID, showExpiry),
		Title:       card.Title,
		Description: description,
		InputMessageContent: &core.InputTextMessageContent{
			MessageText: messageText,
			ParseMode:   "HTML",
		},
	}
}

func inlineResultID(cardID int64, showExpiry bool) string {
	if showExpiry {
		return fmt.Sprintf("card_%d_exp", cardID)
	}
	return fmt.Sprintf("card_%d", cardID)
}

func parseInlineQuery(rawQuery string) (string, bool) {
	rawQuery = strings.TrimSpace(rawQuery)
	if rawQuery == "" {
		return "", false
	}

	parts := strings.Fields(rawQuery)
	if len(parts) == 0 {
		return "", false
	}

	if strings.EqualFold(parts[0], "exp") {
		return strings.TrimSpace(strings.Join(parts[1:], " ")), true
	}

	return rawQuery, false
}
