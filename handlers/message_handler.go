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
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func CommandHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]*models.Account, config models.Config) {
	chatId := update.Message.Chat.ID
	if _, ok := db[chatId]; !ok {
		db[chatId] = &models.Account{"en", make(map[string]bool)}
		SendAdmin(config.AdminChatId, bot, fmt.Sprintf("🤖 I got new user <u>%s (%d)</u>", update.Message.From.UserName, chatId))
		log.Printf("ADD new user: %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
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

	case "adminGetUserNumber":
		if chatId == config.AdminChatId {
			msg.Text = fmt.Sprintf("%d", len(db))
		} else {
			log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
			msg.Text = assets.Texts[locale]["unknown_command"]
		}

	case "adminGetAll":
		if chatId == config.AdminChatId {
			var result string
			for eachUser, igAccounts := range db {
				result = result + fmt.Sprintf("👤 %d:", eachUser)
				if len(igAccounts.IgAccounts) > 0 {
					for igAccount, isPrivate := range igAccounts.IgAccounts {
						statusEmoji := "🟢 "
						if isPrivate {
							statusEmoji = "🔴 "
						}
						result = result + fmt.Sprintf("\n      %s %s", statusEmoji, igAccount)
					}
				} else {
					result = result + " empty"
				}
				result = result + "\n\n"
			}
			msg.Text = result
		} else {
			log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
			msg.Text = assets.Texts[locale]["unknown_command"]
		}

	case "sendAllRu":
		if chatId == config.AdminChatId {
			if len(update.Message.CommandArguments()) > 1 {
				messageToSendAll := update.Message.CommandArguments()
				for eachUser, data := range db {
					if data.Locale == "ru" {
						SendAdmin(eachUser, bot, messageToSendAll)
					}
				}
				msg.Text = "🤖 Message has been sent to all users (ru)"
			}
		} else {
			log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
			msg.Text = assets.Texts[locale]["unknown_command"]
		}

	case "sendAllEn":
		if chatId == config.AdminChatId {
			if len(update.Message.CommandArguments()) > 1 {
				messageToSendAll := update.Message.CommandArguments()
				for eachUser, data := range db {
					if data.Locale == "en" {
						SendAdmin(eachUser, bot, messageToSendAll)
					}
				}
				msg.Text = "🤖 Message has been sent to all users (en)"
			}
		} else {
			log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
			msg.Text = assets.Texts[locale]["unknown_command"]
		}

	default:
		log.Printf("COMMAND UNKNOWN %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
		msg.Text = assets.Texts[locale]["unknown_command"]
	}
	bot.Send(msg)
}

func MessageHandler(workingPath string, bot *tgbotapi.BotAPI, update tgbotapi.Update, db map[int64]*models.Account, config models.Config, insta *goinsta.Instagram) {
	chatId := update.Message.Chat.ID
	messageText := update.Message.Text
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	if _, ok := db[chatId]; !ok {
		db[chatId] = &models.Account{"en", make(map[string]bool)}
	}
	locale := db[chatId].Locale
	if update.Message.From.IsBot {
		msg.Text = assets.Texts[locale]["do_not_work_with_bots"]
		bot.Send(msg)
	} else if messageText == "🌎 English" || messageText == "🇷🇺 Русский" {
		if messageText == "🇷🇺 Русский" {
			locale = "ru"
		} else {
			locale = "en"
		}
		db[chatId].Locale = locale
		utils.SaveDb(db, config)
		msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{}
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
	} else if messageText == assets.Texts[locale]["button_accounts"] {
		accountsOutput := ""
		if len(db[chatId].IgAccounts) == 0 {
			msg.Text = assets.Texts[locale]["account_list_is_empty"]
		} else {
			for eachAccount, isPrivate := range db[chatId].IgAccounts {
				statusEmoji := "🟢 "
				if isPrivate {
					statusEmoji = "🔴 "
				}
				accountsOutput = accountsOutput + statusEmoji + " " + eachAccount + "\n\n"
			}
			msg.Text = accountsOutput
		}
		bot.Send(msg)
	} else if messageText == assets.Texts[locale]["button_delete"] {
		if len(db[chatId].IgAccounts) == 0 {
			msg.Text = assets.Texts[locale]["account_list_is_empty"]
		} else {
			deleteAccountsKeyboard := tgbotapi.InlineKeyboardMarkup{}
			for eachAccount := range db[chatId].IgAccounts {
				var row []tgbotapi.InlineKeyboardButton
				button := tgbotapi.NewInlineKeyboardButtonData(eachAccount, eachAccount)
				row = append(row, button)
				deleteAccountsKeyboard.InlineKeyboard = append(deleteAccountsKeyboard.InlineKeyboard, row)
			}
			msg.Text = assets.Texts[locale]["account_choose_to_delete"]
			msg.ReplyMarkup = deleteAccountsKeyboard
		}
		bot.Send(msg)
	} else if messageText == assets.Texts[locale]["button_help"] {
		msg.ParseMode = "HTML"
		msg.Text = assets.Texts[locale]["instructions"]
		bot.Send(msg)
	} else if strings.Contains(messageText, " ") {
		msg.Text = fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"])
		bot.Send(msg)
	} else if strings.Contains(messageText, "://") {
		go api.DownloadMedia(messageText, workingPath, insta, bot, chatId, locale)
	} else if len(db[chatId].IgAccounts) >= config.SnitchLimit {
		msg.Text = fmt.Sprintf(assets.Texts[locale]["limit_of_accounts"], config.SnitchLimit)
		bot.Send(msg)
	} else {
		newAccountName := strings.ToLower(messageText)
		if newAccountName[0:1] == "@" && len(newAccountName) > 1 {
			newAccountName = utils.TrimFirstChar(newAccountName)
		}
		privateStatus, err := api.GetPrivateStatus(insta, newAccountName)
		if err == api.UserNotFoundError { // ошибка "account_not_found"
			log.Printf("ADD ERROR account not found %s", newAccountName)
			msg.Text = fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"])
			bot.Send(msg)
		} else if _, ok := err.(goinsta.ChallengeError); ok { // TODO разобраться с challenge
			log.Printf("ADD ERROR challenge: %v", err)
			msg.Text = assets.Texts[locale]["panic"]
			bot.Send(msg)
		} else if err != nil { // какая-то другая ошибка
			log.Printf("ADD ERROR %s: %v", newAccountName, err)
			msg.Text = fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"])
			bot.Send(msg)
		} else {
			db[chatId].IgAccounts[newAccountName] = privateStatus
			msg.Text = fmt.Sprintf(assets.Texts[locale]["account_added"], newAccountName)
			msg.ParseMode = "HTML"
			utils.SaveDb(db, config)
			//SendAdmin(config.AdminChatId, bot, fmt.Sprintf("🤖 <u>%s (%d)</u> now tracking for new account", update.Message.From.UserName, update.Message.From.ID))
			SendAdmin(config.AdminChatId, bot, fmt.Sprintf("🤖 <u>%s (%d)</u> now tracking for <u>%s</u>", update.Message.From.UserName, update.Message.From.ID, newAccountName))
			log.Printf("ADD user %s (%d) now tracking for %s", update.Message.From.UserName, update.Message.From.ID, newAccountName)
			bot.Send(msg)
		}
	}
}
