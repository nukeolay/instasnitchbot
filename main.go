package main

import (
	"encoding/json"
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"io/ioutil"
	"log"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ahmdrz/goinsta/v2"
	"github.com/go-co-op/gocron"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	TelegramBotToken   string
	LogFileName        string
	DbName             string
	UseWebhook         bool
	UpdateStatusPeriod int
	UpdateNextAccount  int
	Port               string
}

func getConfig() Config {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	return configuration
}

type account map[string]bool

var db = map[int64]account{}

func saveData(db map[int64]account, config Config) {
	file, err := json.MarshalIndent(db, "", " ")
	ioutil.WriteFile(config.DbName, file, 0644)
	if err != nil {
		log.Printf("SAVE DB ERROR %v", err)
	}
}

func loadData(config Config) {
	file, err := ioutil.ReadFile(config.DbName)
	if err != nil {
		log.Printf("LOAD DB ERROR %v", err)
	} else {
		db = map[int64]account{}
		json.Unmarshal([]byte(file), &db)
	}
}

func taskTryLogin() {

}

func taskUpdateStatus(bot *tgbotapi.BotAPI, insta *goinsta.Instagram, db map[int64]account, igAccounts map[string]string, config Config) {
	time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // пауза перед запуском таска
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
			saveData(db, config)
			time.Sleep(time.Duration(config.UpdateNextAccount * 1000000000)) // проверка следующего аккаунта через _ секунд
		}
	}
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("<html><head><title>InstasnitchBot</title></head><body>Hi there! I'm InstasnitchBot!<br>I can do some shit.<br>You can get me at <a href=\"https://t.me/instasnitchbot\">https://t.me/instasnitchbot</a></body></html>"))
}

func main() {
	isPanic := false
	isTryingToLogin := false
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
			log.Panicf("ERROR ERROR ERROR INSTA IS NIL")
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
	loadData(config)

	//setting up cron update accounts
	s := gocron.NewScheduler(time.UTC)
	_, errS := s.Every(config.UpdateStatusPeriod).Minutes().Do(taskUpdateStatus, bot, insta, db, igAccounts, config)
	if errS != nil {
		log.Printf("CRON ERROR update status %v", errS)
	}
	s.StartAsync()

	if isPanic {

	}

	for update := range updates {
		if isPanic {
			if update.Message == nil { // игнорируем все кроме сообщений
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts["panic"])
			bot.Send(msg)
		} else {
			if update.CallbackQuery != nil { // обработка нажатий на кнопки в телеграме
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))
				var chatId int64 = update.CallbackQuery.Message.Chat.ID
				delete(db[chatId], update.CallbackQuery.Data)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_deleted"], update.CallbackQuery.Data))
				msg.ParseMode = "HTML"
				saveData(db, config)
				bot.Send(msg)
				if len(db[chatId]) == 0 {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, assets.Texts["account_list_is_empty_now"])
					bot.Send(msg)
				}
				continue
			}

			if update.Message == nil { // игнорируем все кроме сообщений
				continue
			}
			var chatId int64 = update.Message.Chat.ID
			if update.Message.Command() != "" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
				switch update.Message.Command() {
				case "start":
					log.Printf("COMMAND /start %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
					msg.ParseMode = "HTML"
					msg.Text = assets.Texts["instructions"]
				case "help":
					log.Printf("COMMAND /help %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
					msg.ParseMode = "HTML"
					msg.Text = assets.Texts["instructions"]
				case "SuperGetUserNumber":
					log.Printf("COMMAND /secret %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
					msg.Text = fmt.Sprintf("%d", len(db))
				case "accounts":
					log.Printf("COMMAND /accounts %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
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
					log.Printf("COMMAND /del %s (ID %d)", update.Message.From.UserName, update.Message.From.ID)
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
			} else {
				// тут добавляем новый аккаунт

				newAccountName := strings.ToLower(update.Message.Text)

				log.Printf("ADD try %s BY %s (ID %d)", newAccountName, update.Message.From.UserName, update.Message.From.ID)
				if _, ok := db[chatId]; !ok {
					db[chatId] = make(account)
				}

				if len(db[chatId]) > 2 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts["limit_of_accounts"])
					bot.Send(msg)
					continue
				}

				if strings.Contains(newAccountName, " ") {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
					bot.Send(msg)
					continue
				}

				if update.Message.From.IsBot {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts["do_not_work_with_bots"])
					bot.Send(msg)
					continue
				}

				//TODO надо выработать единые правила для нулевой инсты - брать паузу для логина (как-то через крон) или выключать бота
				//TODO скорее всего этот блок нужно удалить, потому, что insta с нулем сюда не сможет попасть (или сможет?)
				//TODO скорее всего не сможет, потому что на старте происход проверка, если после загрузки ноль, то паника и бот выключается
				// if insta == nil {
				// 	log.Printf("ADD error %s", newAccountName)
				// 	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["endpoint_error"]))
				// 	bot.Send(msg)
				////isPanic = true
				// 	continue
				// }

				privateStatus, err := api.GetPrivateStatus(insta, newAccountName)
				if err == api.UserNotFoundError { // ошибка "account_not_found"
					log.Printf("ADD error %s, %v", newAccountName, err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
					bot.Send(msg)
					continue
				} else if err != nil { // ошибка при проверке статуса кроме "account_not_found"
					log.Printf("ADD ERROR %s, %v", newAccountName, err)
					insta = api.GetNewApi(igAccounts)
					if insta == nil { // ошибка авторизации
						//TODO если инста нулевая, то потом может произойти что угодно,
						//TODO тут надо прекращать таск и брать паузу для логина (как-то через крон) или выключать бота
						////isPanic = true
						log.Print("ADD ERROR login")
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], err.Error()))
						bot.Send(msg)
						continue
					} else { // авторизация прошла успешно
						privateStatus, err = api.GetPrivateStatus(insta, newAccountName)
						if err == api.UserNotFoundError { // ошибка "account_not_found"
							log.Printf("ADD error %s, %v", newAccountName, err)
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
							bot.Send(msg)
						} else if err != nil {
							log.Print("ADD ERROR login")
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], err.Error()))
							bot.Send(msg)
							continue
						}
					}
				}

				db[chatId][newAccountName] = privateStatus
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_added"], newAccountName))
				log.Printf("ADD success %s BY %s (ID %d)", newAccountName, update.Message.From.UserName, update.Message.From.ID)
				msg.ParseMode = "HTML"
				saveData(db, config)
				bot.Send(msg)
			}

		}

	}
}
