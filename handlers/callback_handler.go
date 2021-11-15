package handlers

import (
	"fmt"
	"instasnitchbot/assets"
	"instasnitchbot/models"
	"instasnitchbot/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)



func CallBackHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]*models.Account, config models.Config) {
	bot.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
	var chatId int64 = update.CallbackQuery.Message.Chat.ID
	locale := db[chatId].Locale
	delete(db[chatId].IgAccounts, update.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["account_deleted"], update.CallbackQuery.Data))
	msg.ParseMode = "HTML"
	utils.SaveDb(db, config)
	bot.Send(msg)
	if len(db[chatId].IgAccounts) == 0 {
		msg.Text = assets.Texts[locale]["account_list_is_empty_now"]
		bot.Send(msg)
	}
}