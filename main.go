package main

import (
	"log"
	"net/http"
	"os"
	"time"

	cb "share_card_robot/callback_query"
	"share_card_robot/commands"
	"share_card_robot/config"
	"share_card_robot/core"
	"share_card_robot/database"
	"share_card_robot/middlewares"
	"share_card_robot/services"
	"share_card_robot/states"
)

func main() {
	http.DefaultClient.Timeout = 60 * time.Second

	if err := config.LoadDotEnv(".env"); err != nil {
		panic("Failed to load .env: " + err.Error())
	}

	if err := database.InitConnection(); err != nil {
		panic("Failed to connect to the database: " + err.Error())
	}
	defer database.CloseConnection()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		panic("BOT_TOKEN muhit o'zgaruvchisi kiritilmagan")
	}

	botUsername := os.Getenv("BOT_USERNAME")
	if botUsername == "" {
		botUsername = "share_card_robot"
	}

	bot := core.BotInit(token, botUsername)
	bot.AddMiddleware(core.ErrorHandlingMiddleware)
	bot.AddMiddleware(core.AuthMiddleware)
	bot.AddMiddleware(core.TimingMiddleware)
	{
		bot.HandleCommand("/start", commands.StartHandler, middlewares.RequireCurrentUser)
		bot.HandleCommand("/help", commands.HelpHandler, middlewares.RequireCurrentUser)
		bot.HandleCommand("/add_card", commands.AddCardHandler, middlewares.RequireCurrentUser)
		bot.HandleCommand("/my_cards", commands.MyCardsHandler, middlewares.RequireCurrentUser)
		bot.HandleCommand("/about", commands.AboutHandler, middlewares.RequireCurrentUser)
	}
	{
		bot.HandleCommand("💳 Mening kartalarim", commands.MyCardsHandler, middlewares.RequireCurrentUser)
		bot.HandleCommand("➕ Karta qo'shish", commands.AddCardHandler, middlewares.RequireCurrentUser)
	}
	{
		bot.HandleState("add_card_name", states.AddCardNameState, middlewares.RequireCurrentUser)
		bot.HandleState("add_card_number", states.AddCardNumberState, middlewares.RequireCurrentUser)
		bot.HandleState("add_card_expiry", states.AddCardExpiryState, middlewares.RequireCurrentUser)
	}
	{
		bot.HandleCallback(services.CallbackDeleteCardPrompt, cb.DeleteCardPromptCallback, middlewares.RequireCurrentUserCallback)
		bot.HandleCallback(services.CallbackDeleteCardConfirm, cb.DeleteCardConfirmCallback, middlewares.RequireCurrentUserCallback)
		bot.HandleCallback(services.CallbackDeleteCardCancel, cb.DeleteCardCancelCallback, middlewares.RequireCurrentUserCallback)
		bot.HandleCallback(services.CallbackCardsPage, cb.CardsPageCallback, middlewares.RequireCurrentUserCallback)
	}
	{
		bot.HandleInlineQuery(commands.InlineSearchHandler, middlewares.RequireCurrentUserInline, middlewares.RequireUserHasCardsInline)
		bot.HandleChosenInlineResult(commands.ChosenInlineResultHandler, middlewares.RequireCurrentUserChosenInline)
	}
	bot.OnQuery(services.OnQuery, commands.DefaultHandler, middlewares.RequireCurrentUser)

	if err := bot.StartServer(); err != nil {
		log.Fatal(err)
	}
}
