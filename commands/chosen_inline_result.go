package commands

import (
	"log"
	"strings"

	"share_card_robot/core"
	"share_card_robot/database/models"
)

func ChosenInlineResultHandler(ctx *core.ChosenInlineResult) {
	var userID int64
	if ctx.AuthUser != nil {
		userID = ctx.AuthUser.ID
	} else if ctx.From != nil {
		user, err := models.GetUserByTelegramID(ctx.From.ID)
		if err != nil {
			log.Printf("chosen inline lookup error: %v", err)
			return
		}
		if user != nil {
			userID = user.ID
		}
	}

	if userID == 0 {
		return
	}

	if err := models.IncrementInlineSearchResultChoice(ctx.ResultID, strings.TrimSpace(ctx.Query), userID); err != nil {
		log.Printf("chosen_inline_result persist error: result_id=%s err=%v", ctx.ResultID, err)
	}
}
