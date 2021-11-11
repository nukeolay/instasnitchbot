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

	"github.com/Davincible/goinsta"
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

func CallBackHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]*models.Account, config models.Config) {
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
	var chatId int64 = update.CallbackQuery.Message.Chat.ID
	locale := db[chatId].Locale
	delete(db[chatId].IgAccounts, update.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["account_deleted"], update.CallbackQuery.Data))
	msg.ParseMode = "HTML"
	utils.SaveDb(db, config)
	bot.Send(msg)
	if len(db[chatId].IgAccounts) == 0 {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, assets.Texts[locale]["account_list_is_empty_now"])
		bot.Send(msg)
	}
}

func CommandHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]*models.Account, config models.Config) {
	chatId := update.Message.Chat.ID
	if _, ok := db[chatId]; !ok {
		db[chatId] = &models.Account{"en", make(map[string]bool)}
	}
	locale := db[chatId].Locale
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	switch update.Message.Command() {
	case "start":
		var chooseLocaleKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🌎 English"),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🇷🇺 Русский"),
			),
		)
		// выбираем язык для первого сообщения на основании локали пользователя
		if update.Message.From.LanguageCode == "ru" {
			msg.Text = assets.Texts["ru"]["choose_language"]
		} else {
			msg.Text = assets.Texts["en"]["choose_language"]
		}
		msg.ReplyMarkup = chooseLocaleKeyboard
	case "help":
		msg.ParseMode = "HTML"
		msg.Text = assets.Texts[locale]["instructions"]
	case "AdminGetUserNumber":
		if chatId == config.AdminChatId {
			msg.Text = fmt.Sprintf("%d", len(db))
		} else {
			log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
			msg.Text = assets.Texts[locale]["unknown_command"]
		}
	case "SendBroadcast":
		if chatId == config.AdminChatId {
			//TODO цикл в котором перебирать всех пользователей и отправлять сообщения (возможно асинхронно, через go)
			//TODO отдельная функция BroadCast(db, message), из doc CommandArguments
		} else {
			log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
			msg.Text = assets.Texts[locale]["unknown_command"]
		}
	case "accounts":
		accountsOutput := ""
		for eachAccount, isPrivate := range db[chatId].IgAccounts {
			statusEmoji := "🟢 "
			if isPrivate {
				statusEmoji = "🔴 "
			}
			accountsOutput = accountsOutput + statusEmoji + " " + eachAccount + "\n\n"
		}
		if accountsOutput == "" {
			msg.Text = assets.Texts[locale]["account_list_is_empty"]
		} else {
			msg.Text = accountsOutput
		}
	case "del":
		deleteAccountsKeyboard := tgbotapi.InlineKeyboardMarkup{}
		for eachAccount := range db[chatId].IgAccounts {
			var row []tgbotapi.InlineKeyboardButton
			button := tgbotapi.NewInlineKeyboardButtonData(eachAccount, eachAccount)
			row = append(row, button)
			deleteAccountsKeyboard.InlineKeyboard = append(deleteAccountsKeyboard.InlineKeyboard, row)
		}
		msg.Text = assets.Texts[locale]["account_choose_to_delete"]
		msg.ReplyMarkup = deleteAccountsKeyboard
	default:
		log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		msg.Text = assets.Texts[locale]["unknown_command"]
	}
	bot.Send(msg)
}

func MessageHandler(workingPath string, bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]*models.Account, config models.Config, insta *goinsta.Instagram) {
	chatId := update.Message.Chat.ID
	messageText := update.Message.Text
	if _, ok := db[chatId]; !ok {
		db[chatId] = &models.Account{"en", make(map[string]bool)}
	}
	locale := db[chatId].Locale
	if update.Message.From.IsBot {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts[locale]["do_not_work_with_bots"])
		bot.Send(msg)
	} else if messageText == "🌎 English" || messageText == "🇷🇺 Русский" {
		if messageText == "🇷🇺 Русский" {
			locale = "ru"
		} else {
			locale = "en"
		}
		db[chatId].Locale = locale
		utils.SaveDb(db, config)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardHide{HideKeyboard: true}

		var standardMenuKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(assets.Texts[locale]["button_accounts"]),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(assets.Texts[locale]["button_delete"]),
			),
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(assets.Texts[locale]["button_help"]),
			),
		)
		msg.ReplyMarkup = standardMenuKeyboard
		msg.ParseMode = "HTML"
		msg.Text = assets.Texts[locale]["instructions"]
		bot.Send(msg)
	} else if strings.Contains(messageText, " ") {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"]))
		bot.Send(msg)
	} else if strings.Contains(messageText, "://") {
		api.DownloadMedia(messageText, workingPath, insta, bot, chatId, locale)
	} else if len(db[chatId].IgAccounts) >= config.SnitchLimit {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["limit_of_accounts"], config.SnitchLimit))
		bot.Send(msg)
	} else {
		newAccountName := strings.ToLower(messageText)
		privateStatus, err := api.GetPrivateStatus(insta, newAccountName)
		if err == api.UserNotFoundError { // ошибка "account_not_found"
			log.Printf("ADD ERROR account not found %s: %v", newAccountName, err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"]))
			bot.Send(msg)
		} else if _, ok := err.(goinsta.ChallengeError); ok { // TODO разобраться с challenge
			log.Printf("ADD ERROR challenge: %v", err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts[locale]["panic"])
			bot.Send(msg)
		} else if err != nil { // какая-то другая ошибка
			log.Printf("ADD ERROR %s: %v", newAccountName, err)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"]))
			bot.Send(msg)
		} else {
			db[chatId].IgAccounts[newAccountName] = privateStatus
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts[locale]["account_added"], newAccountName))
			msg.ParseMode = "HTML"
			utils.SaveDb(db, config)
			SendAdmin(config.AdminChatId, bot, fmt.Sprintf("<u>%s (%d)</u> added <u>%s</u>", update.Message.From.UserName, update.Message.From.ID, newAccountName))
			bot.Send(msg)
		}
	}
}
