package main

import (
	"encoding/json"
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"instasnitchbot/handlers"
	"instasnitchbot/models"
	"instasnitchbot/utils"

	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/go-co-op/gocron"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getConfig() models.Config {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := models.Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	return configuration
}

func taskTryLogin(igAccounts map[string]string) *goinsta.Instagram {
	insta, errLoad := goinsta.Import(".goinsta")
	if errLoad != nil { // если ошибка импорта
		log.Printf("INSTA ERROR import: %v", errLoad)
		return insta // возвращаю или ноль, или залогиненный
	}
	log.Print("INSTA import success")
	return insta // возвращаю залогиненный по импорту
}

func taskUpdateStatus(bot *tgbotapi.BotAPI, insta *goinsta.Instagram, db map[int64]models.Account, igAccounts map[string]string, config models.Config) {
	time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // пауза перед запуском таска
	//TODO проверять insta на нуль, если нуль, то завершать и запускать функцию taskTryLogin

	log.Printf("CRON started")
	for chatId, storedAccounts := range db {
		for accountName, oldPrivateStatus := range storedAccounts {
			log.Printf("CRON updating %s", accountName)
			newPrivateStatus, err := api.GetPrivateStatus(insta, strings.ToLower(accountName))
			if err == api.UserNotFoundError { // ошибка "account_not_found"
				continue
			} else if err != nil { // ошибка при проверке статуса кроме "account_not_found"
				log.Printf("CRON ERROR updating %s, %v", accountName, err)
				insta = api.GetNewApi(igAccounts)
				if insta == nil { // ошибка авторизации
					//TODO надо выработать единые правила для нулевой инсты - брать паузу для логина (как-то через крон) или выключать бота
					//TODO если инста нулевая, то потом может произойти что угодно,
					//TODO тут надо прекращать таск и брать паузу для логина (как-то через крон) или выключать бота
					log.Print("CRON ERROR login")
					continue
				} else { // авторизация прошла успешно
					newPrivateStatus, err = api.GetPrivateStatus(insta, strings.ToLower(accountName))
					if err != nil { // это ошибка не связанная с логином, возможно поменялось имя акканта, пропустить обновление
						log.Printf("CRON ERROR updating %s, %v", accountName, err)
						continue
					}
				}
			}
			if newPrivateStatus != oldPrivateStatus { // если статус приватности изменился, то отправляем сообщение
				msg := tgbotapi.NewMessage(chatId, "")
				db[chatId][accountName] = newPrivateStatus // записываем в db новый статус
				if newPrivateStatus {
					msg.Text = fmt.Sprintf(assets.Texts["account_is_private"], accountName)
				} else {
					msg.Text = fmt.Sprintf(assets.Texts["account_is_not_private"], accountName)
				}
				log.Printf("CRON %s status updated", accountName)
				msg.ParseMode = "HTML"
				bot.Send(msg)
			}
			utils.SaveDb(db, config)
			time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // проверка следующего аккаунта через _ секунд
		}
	}
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("<html><head><title>InstasnitchBot</title></head><body>Hi there! I'm InstasnitchBot!<br>I can do some shit.<br>You can get me at <a href=\"https://t.me/instasnitchbot\">https://t.me/instasnitchbot</a></body></html>"))
}

func main() {
	// initialazing
	// isPanic := false
	// isTryingToLogin := false
	config := getConfig()
	port := config.Port
	if config.Port == "" {
		port = os.Getenv("PORT")
	}
	http.HandleFunc("/", MainHandler)
	go http.ListenAndServe(":"+port, nil)

	// setting up log
	f, err := os.OpenFile(config.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("ERROR opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// instagram login
	igAccounts := api.LoadLogins()
	insta := api.GetSavedApi(igAccounts)
	if insta == nil { // не получилось импортировать
		insta = api.GetNewApi(igAccounts)
		if insta == nil { // не получилось залигиниться
			//TODO надо выработать единые правила для нулевой инсты - брать паузу для логина (как-то через крон) или выключать бота
			////isPanic = true
			log.Panic("ERROR ERROR ERROR INSTA IS NIL")
		}
	}

	//setting up bot
	bot, err := tgbotapi.NewBotAPI(config.TelegramBotToken)
	if err != nil {
		log.Panic(err)
	}
	log.Println("--------------------------BOT STARTED--------------------------")
	bot.Debug = false

	//setting up bot telegram connection
	updates := bot.ListenForWebhook("/" + bot.Token)
	if !config.UseWebhook {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates, _ = bot.GetUpdatesChan(u)
	}
	db := utils.LoadDb(config)

	//setting up cron update accounts
	s := gocron.NewScheduler(time.UTC)
	_, errS := s.Every(config.UpdateStatusPeriod).Minutes().Do(taskUpdateStatus, bot, insta, db, igAccounts, config)
	if errS != nil {
		log.Printf("CRON ERROR update status %v", errS)
	}
	if insta == nil { // если на старте не получилось залогиниться, то...

	} else { // если на старте залогинились, то...
		s.StartAsync()
	}

	for update := range updates {
		// если инста ноль
		if insta == nil {
			if update.Message == nil { // игнорируем все кроме сообщений
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts["panic"])
			bot.Send(msg)
			continue
		}

		if update.CallbackQuery != nil { // обработка нажатий на кнопки в телеграме
			handlers.CallBackHandler(bot, update, db, config)
			continue
		} else if update.Message == nil { // игнорируем все кроме сообщений
			continue
		} else if update.Message.Command() != "" { // обработка сообщений
			handlers.CommandHandler(bot, update, db)
			continue
		} else { // добавляем новый аккаунт
			handlers.AddNewSnitch(bot, update, db, config, insta)
			continue
		}
	}
}
