package main

import (
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"instasnitchbot/handlers"
	"instasnitchbot/models"
	"instasnitchbot/utils"
	"path"
	"runtime"

	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Davincible/goinsta"
	"github.com/go-co-op/gocron"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func taskStatusUpdater(bot *tgbotapi.BotAPI, insta **goinsta.Instagram, db map[int64]*models.Account, igAccounts map[string]string, config models.Config, loginCountdown *int, isTaskFinished *bool) {
	if *isTaskFinished {
		*isTaskFinished = false
		time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // пауза перед запуском таска
		log.Printf("CRON started")
		// если крон начал обновлять статусы и дел, что инста нуль,
		if *insta == nil {
			log.Printf("CRON ERROR insta is nil")
			handlers.SendAdmin(config.AdminChatId, bot, "CRON ERROR insta is nil")
			// проверяет переменную loginCountdown, если она 0, значит либо еще не запускалась,
			// либо прошло config.TryLoginPeriod (5) циклов обновления по 10 минут (как настроить)
			// значит пора обновлять снова
			if *loginCountdown == config.TryLoginPeriod {
				*loginCountdown = 0
			}
			if *loginCountdown == 0 {
				log.Printf("CRON ERROR trying to login")
				handlers.SendAdmin(config.AdminChatId, bot, "CRON ERROR trying to login")
				*insta = api.GetNewApi(igAccounts)

			}
			*loginCountdown++ // в каждом цикле прибавляем 1
		} else {
			// если инста не нуль, то сбрасываем счетчик loginCountdown на 0
			*loginCountdown = 0
			for chatId, storedAccounts := range db {
				for accountName, oldPrivateStatus := range storedAccounts.IgAccounts {
					locale := db[chatId].Locale
					log.Printf("CRON updating %s", accountName)
					newPrivateStatus, err := api.GetPrivateStatus(*insta, strings.ToLower(accountName))
					if err == api.UserNotFoundError { // ошибка "account_not_found"
						continue
					} else if _, ok := err.(goinsta.ChallengeError); ok { // TODO разобраться с challenge
						//TODO отправлять мне в телеграм ошибку
						log.Printf("CRON ERROR challenge: %v", err)
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

func main() {
	// initializing
	loginCountdown := 0
	isTaskFinished := true
	config := utils.GetConfig()
	port := config.Port
	if config.Port == "" {
		port = os.Getenv("PORT")
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	workingPath := path.Dir(filename)

	http.HandleFunc("/", handlers.WebHandler)
	go http.ListenAndServe(":"+port, nil)

	// setting up log
	f, err := os.OpenFile(config.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("START ERROR opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Print("--------------------------INITIALIZED--------------------------")
	log.Print("----------------INSTAGRAM LOGIN----------------")
	igAccounts := api.LoadLogins()
	insta := api.GetSavedApi(igAccounts)

	if insta == nil { // не получилось импортировать
		insta = api.GetNewApi(igAccounts)
		if insta == nil { // не получилось залигиниться
			log.Panic("START ERROR insta is nil")
		}
	}

	//setting up bot
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	log.Println("----------------BOT STARTED----------------")
	bot.Debug = false

	//setting up bot telegram connection
	updates := bot.ListenForWebhook("/" + bot.Token)
	if !config.UseWebhook {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = bot.GetUpdatesChan(u)
	}
	db := utils.LoadDb(config)

	//setting up cron update account
	log.Println("----------------SETTING UP CRON----------------")
	cronStatusUpdater := gocron.NewScheduler(time.UTC)
	_, errCronStatusUpdater := cronStatusUpdater.Every(config.UpdateStatusPeriod).Minutes().Do(taskStatusUpdater, bot, &insta, db, igAccounts, config, &loginCountdown, &isTaskFinished)
	if errCronStatusUpdater != nil {
		log.Printf("START CRON ERROR: %v", errCronStatusUpdater)
		handlers.SendAdmin(config.AdminChatId, bot, fmt.Sprintf("START CRON ERROR: %v", errCronStatusUpdater))
	} else {
		cronStatusUpdater.StartAsync()
	}

	//-----------------------------------HANDLING UPDATES-----------------------------------//
	for update := range updates {

		// если инста ноль
		if insta == nil {
			if update.Message == nil { // игнорируем все кроме сообщений
				continue
			}
			locale := db[update.Message.Chat.ID].Locale
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts[locale]["panic"])
			bot.Send(msg)
			continue
		}

		if update.CallbackQuery != nil { // обработка нажатий на кнопки в телеграме
			handlers.CallBackHandler(bot, update, db, config)
			continue
		} else if update.Message == nil { // игнорируем все кроме сообщений
			continue
		} else if update.Message.Command() != "" { // обработка сообщений
			handlers.CommandHandler(bot, update, db, config)
			continue
		} else { // добавляем новый аккаунт
			handlers.MessageHandler(workingPath, bot, update, db, config, insta)
			continue
		}
	}
}
