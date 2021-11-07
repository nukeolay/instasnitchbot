package handlers

import (
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"instasnitchbot/models"
	"instasnitchbot/utils"
	"log"
	"net/http"
	"strings"

	"github.com/ahmdrz/goinsta/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func WebHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("<html><head><title>InstasnitchBot</title></head><body>Hi there! I'm InstasnitchBot!<br>I can do some shit.<br>You can get me at <a href=\"https://t.me/instasnitchbot\">https://t.me/instasnitchbot</a></body></html>"))
}

func SendAdmin(chatId int64, bot *tgbotapi.BotAPI, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

func CallBackHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]models.Account, config models.Config) {
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
	var chatId int64 = update.CallbackQuery.Message.Chat.ID
	delete(db[chatId], update.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_deleted"], update.CallbackQuery.Data))
	msg.ParseMode = "HTML"
	utils.SaveDb(db, config)
	bot.Send(msg)
	if len(db[chatId]) == 0 {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, assets.Texts["account_list_is_empty_now"])
		bot.Send(msg)
	}
}

func CommandHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]models.Account) {
	chatId := update.Message.Chat.ID
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	switch update.Message.Command() {
	case "start":
		////log.Printf("COMMAND /start %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		msg.ParseMode = "HTML"
		msg.Text = assets.Texts["instructions"]
	case "help":
		////log.Printf("COMMAND /help %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		msg.ParseMode = "HTML"
		msg.Text = assets.Texts["instructions"]
	case "SuperGetUserNumber":
		////log.Printf("COMMAND /secret %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		msg.Text = fmt.Sprintf("%d", len(db))
	case "accounts":
		////log.Printf("COMMAND /accounts %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		accountsOutput := ""
		for eachAccount, isPrivate := range db[chatId] {
			statusEmoji := "🟢 "
			if isPrivate {
				statusEmoji = "🔴 "
			}
			accountsOutput = accountsOutput + statusEmoji + " " + eachAccount + "\n\n"
		}
		if accountsOutput == "" {
			msg.Text = assets.Texts["account_list_is_empty"]
		} else {
			msg.Text = accountsOutput
		}
	case "del":
		////log.Printf("COMMAND /del %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		deleteAccountsKeyboard := tgbotapi.InlineKeyboardMarkup{}
		for eachAccount := range db[chatId] {
			var row []tgbotapi.InlineKeyboardButton
			button := tgbotapi.NewInlineKeyboardButtonData(eachAccount, eachAccount)
			row = append(row, button)
			deleteAccountsKeyboard.InlineKeyboard = append(deleteAccountsKeyboard.InlineKeyboard, row)
		}

		msg.Text = assets.Texts["account_choose_to_delete"]
		msg.ReplyMarkup = deleteAccountsKeyboard
	default:
		log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		msg.Text = assets.Texts["unknown_command"]
	}
	bot.Send(msg)
}

func MessageHandler(workingPath string, bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]models.Account, config models.Config, insta *goinsta.Instagram) {
	chatId := update.Message.Chat.ID
	messageText := update.Message.Text

	////log.Printf("ADD try %s BY %s (ID %d)", newAccountName, update.Message.From.UserName, update.Message.From.ID)
	if _, ok := db[chatId]; !ok {
		db[chatId] = make(models.Account)
	}

	if strings.Contains(messageText, " ") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
		bot.Send(msg)
	} else if update.Message.From.IsBot {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts["do_not_work_with_bots"])
		bot.Send(msg)
	} else if strings.Contains(messageText, "://") {
		api.DownloadMedia(messageText, workingPath, insta, bot, chatId)
	} else if len(db[chatId]) >= config.SnitchLimit {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["limit_of_accounts"], config.SnitchLimit))
		bot.Send(msg)
	} else {
		newAccountName := strings.ToLower(messageText)
		privateStatus, err := api.GetPrivateStatus(insta, newAccountName)
		if err == api.UserNotFoundError { // ошибка "account_not_found"
			log.Printf("ADD error %s, %v", newAccountName, err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
			bot.Send(msg)
		} else if err != nil { // какая-то другая ошибка
			log.Printf("ADD ERROR %s, %v", newAccountName, err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
			bot.Send(msg)
		} else {
			db[chatId][newAccountName] = privateStatus
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_added"], newAccountName))
			////log.Printf("ADD success %s BY %s (ID %d)", newAccountName, update.Message.From.UserName, update.Message.From.ID)
			msg.ParseMode = "HTML"
			utils.SaveDb(db, config)
			SendAdmin(config.AdminChatId, bot, fmt.Sprintf("<u>%s</u> added by <u>%s</u> (ID <u>%d</u>)", newAccountName, update.Message.From.UserName, update.Message.From.ID))
			bot.Send(msg)
		}
	}
}
