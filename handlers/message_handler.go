package handlers

import (
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"instasnitchbot/models"
	"instasnitchbot/utils"
	"log"
	"strings"

	"github.com/Davincible/goinsta"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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
		if len(newAccountName) == 0 {
			return
		}
		if newAccountName[0:1] == "@" && len(newAccountName) > 1 {
			newAccountName = utils.TrimFirstChar(newAccountName)
		}
		privateStatus, err := api.GetPrivateStatus(insta, newAccountName)
		if err == api.UserNotFoundError { // error "account_not_found"
			log.Printf("ADD ERROR account not found %s", newAccountName)
			msg.Text = fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"])
			bot.Send(msg)
		} else if _, ok := err.(goinsta.ChallengeError); ok { // TODO make challenge handler
			log.Printf("ADD ERROR challenge: %v", err)
			SendAdmin(config.AdminChatId, bot, fmt.Sprintf("ADD ERROR challenge: %v", err))
			msg.Text = assets.Texts[locale]["panic"]
			bot.Send(msg)
		} else if err != nil { // any other error
			log.Printf("ADD ERROR %s: %v", newAccountName, err.Error()[0:10])
			msg.Text = fmt.Sprintf(assets.Texts[locale]["account_add_error"], assets.Texts[locale]["account_not_found"])
			bot.Send(msg)
		} else {
			db[chatId].IgAccounts[newAccountName] = privateStatus
			if privateStatus {
				msg.Text = fmt.Sprintf(assets.Texts[locale]["account_added"], newAccountName)
			} else {
				msg.Text = fmt.Sprintf(assets.Texts[locale]["account_added_not_private"], newAccountName)
			}
			msg.ParseMode = "HTML"
			utils.SaveDb(db, config)
			SendAdmin(config.AdminChatId, bot, fmt.Sprintf("🤖 <u>%s (%d)</u> now tracking for <u>%s</u>", update.Message.From.UserName, update.Message.From.ID, newAccountName))
			log.Printf("ADD user %s (%d) now tracking for %s", update.Message.From.UserName, update.Message.From.ID, newAccountName)
			bot.Send(msg)
		}
	}
}
