package handlers

import (
	"fmt"
	"log"
	"instasnitchbot/assets"
	"instasnitchbot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
				result = result + fmt.Sprintf("👤 %d [%s]:", eachUser, igAccounts.Locale)
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

	case "del":
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