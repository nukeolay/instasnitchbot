package handlers

import (
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"instasnitchbot/models"
	"instasnitchbot/utils"
	"log"
	"strings"
	"time"

	"github.com/Davincible/goinsta"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


func TaskStatusUpdater(bot *tgbotapi.BotAPI, insta **goinsta.Instagram, db map[int64]*models.Account, igAccounts map[string]string, config models.Config, loginCountdown *int, isTaskFinished *bool) {
	if *isTaskFinished {
		*isTaskFinished = false
		time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // пауза перед запуском таска
		log.Printf("CRON started")
		// если крон начал обновлять статусы и дел, что инста нуль,
		if *insta == nil {
			log.Printf("CRON ERROR insta is nil")
			SendAdmin(config.AdminChatId, bot, "CRON ERROR insta is nil")
			// проверяет переменную loginCountdown, если она 0, значит либо еще не запускалась,
			// либо прошло config.TryLoginPeriod (5) циклов обновления по 10 минут (как настроить)
			// значит пора обновлять снова
			if *loginCountdown == config.TryLoginPeriod {
				*loginCountdown = 0
			}
			if *loginCountdown == 0 {
				log.Printf("CRON ERROR trying to login")
				SendAdmin(config.AdminChatId, bot, "CRON ERROR trying to login")
				*insta = api.GetNewApi(igAccounts)

			}
			*loginCountdown++ // в каждом цикле прибавляем 1
		} else {
			// если инста не нуль, то сбрасываем счетчик loginCountdown на 0
			*loginCountdown = 0
			for chatId, storedAccounts := range db {
				for accountName, oldPrivateStatus := range storedAccounts.IgAccounts {
					locale := db[chatId].Locale
					//log.Printf("CRON updating %s", accountName)
					newPrivateStatus, err := api.GetPrivateStatus(*insta, strings.ToLower(accountName))
					if err == api.UserNotFoundError { // ошибка "account_not_found"
						continue
					} else if _, ok := err.(goinsta.ChallengeError); ok { // TODO разобраться с challenge
						log.Printf("CRON ERROR challenge: %v", err)
						SendAdmin(config.AdminChatId, bot, fmt.Sprintf("CRON ERROR challenge: %v", err))
					} else if err != nil { // ошибка при проверке статуса кроме "account_not_found" и "challenge"
						log.Printf("CRON ERROR updating %s: %v", accountName, err)
					} else {
						if newPrivateStatus != oldPrivateStatus { // если статус приватности изменился, то отправляем сообщение
							msg := tgbotapi.NewMessage(chatId, "")
							db[chatId].IgAccounts[accountName] = newPrivateStatus // записываем в db новый статус
							if newPrivateStatus {
								msg.Text = fmt.Sprintf(assets.Texts[locale]["account_is_private"], accountName)
							} else {
								msg.Text = fmt.Sprintf(assets.Texts[locale]["account_is_not_private"], accountName)
							}
							log.Printf("CRON %s status updated", accountName)
							msg.ParseMode = "HTML"
							bot.Send(msg)
						}
						utils.SaveDb(db, config)
					}
					time.Sleep(time.Duration(utils.GetRandomUpdateNextAccount(config.UpdateNextAccount) * 1000000000)) // проверка следующего аккаунта через _ секунд (получаем случайное число)
				}
			}
		}
		*isTaskFinished = true
	} else {
		log.Printf("CRON ERROR task conflict, but don't worry, isTaskFinished works well")
	}
}