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
		time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // pause before next task
		log.Printf("CRON started")
		// if cron is working now and insta is nil
		if *insta == nil {
			log.Printf("CRON ERROR insta is nil")
			SendAdmin(config.AdminChatId, bot, "CRON ERROR insta is nil")
			// check loginCountdown, if it equals 0 it means it was not started yet ot passed config.TryLoginPeriod (5) cycles
			if *loginCountdown == config.TryLoginPeriod {
				*loginCountdown = 0
			}
			if *loginCountdown == 0 {
				log.Printf("CRON ERROR trying to login")
				SendAdmin(config.AdminChatId, bot, "CRON ERROR trying to login")
				*insta = api.GetNewApi(igAccounts)

			}
			*loginCountdown++ // every cycle adds 1
		} else {
			// if insta is not nil, reset counter loginCountdown to 0
			*loginCountdown = 0
			for chatId, storedAccounts := range db {
				for accountName, oldPrivateStatus := range storedAccounts.IgAccounts {
					locale := db[chatId].Locale
					//log.Printf("CRON updating %s", accountName)
					newPrivateStatus, err := api.GetPrivateStatus(*insta, strings.ToLower(accountName))
					if err == api.UserNotFoundError { // error "account_not_found"
						continue
					} else if _, ok := err.(goinsta.ChallengeError); ok { // TODO make challenge handler
						log.Printf("CRON ERROR challenge: %v", err)
						SendAdmin(config.AdminChatId, bot, fmt.Sprintf("CRON ERROR challenge: %v", err))
					} else if err != nil { // error while checking privacy status except "account_not_found" and "challenge"
						log.Printf("CRON ERROR updating %s: %v", accountName, err)
					} else {
						if newPrivateStatus != oldPrivateStatus { // if privacy status changed send message
							msg := tgbotapi.NewMessage(chatId, "")
							db[chatId].IgAccounts[accountName] = newPrivateStatus // update db with new privacy status
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
					time.Sleep(time.Duration(utils.GetRandomUpdateNextAccount(config.UpdateNextAccount) * 1000000000)) // next account will be checked in _ seconds (random time)
				}
			}
		}
		*isTaskFinished = true
	} else {
		log.Printf("CRON ERROR task conflict, but don't worry, isTaskFinished works well")
	}
}
