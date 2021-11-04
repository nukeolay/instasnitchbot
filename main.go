package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"io/ioutil"
	"log"

	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-co-op/gocron"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	TelegramBotToken string
	LogFileName      string
	DbName           string
	UseWebhook       bool
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

var updatePeriod = 20             // запуск cron каждые __ минут
var updateStep = 40 * time.Second // шаг с которым обновляются аккаунты в каждом цикле cron

type account map[string]bool

var db = map[int64]account{}

func getPrivateStatus(accountName string) (isPrivate bool, err error) {
	isPrivate, err = api.GetPrivateStatusTopSearch(accountName)
	if _, ok := err.(api.EndpointErrorParsing); !ok {
		if err != nil { // какая-то ошибка
			if _, ok := err.(api.EndpointErrorAccountNotFound); ok { // нет такого аккаунта
				return true, errors.New(assets.Texts["account_not_found"])
			}
			return true, err // какая-то ошибка, не дело не в блокировке endpoint
		} else {
			return isPrivate, nil // успех!
		}
	} else {
		log.Printf("ENDPOINT ERROR topsearch blocked")
		// если первый endpoint заблокирован, то пробуем по второму
		isPrivate, err := api.GetPrivateStatusA1Channel(accountName)
		if _, ok := err.(api.EndpointErrorParsing); !ok {
			if err != nil { // какая-то ошибка
				if _, ok := err.(api.EndpointErrorAccountNotFound); ok { // нет такого аккаунта
					return true, errors.New(assets.Texts["account_not_found"])
				}
				return true, err // какая-то ошибка, не дело не в блокировке endpoint
			} else {
				return isPrivate, nil // успех!
			}
		} else {
			log.Printf("ENDPOINT ERROR a1/channel blocked")
			// если второй endpoint заблокирован, то пробуем по третьему
			isPrivate, err := api.GetPrivateStatusA1(accountName)
			if _, ok := err.(api.EndpointErrorParsing); ok {
				log.Printf("ENDPOINT ERROR a1 blocked")
				return true, errors.New(assets.Texts["endpoint_error"]) // endpoint заблокирован
			}
			if err != nil { // какая-то ошибка
				if _, ok := err.(api.EndpointErrorAccountNotFound); ok { // нет такого аккаунта
					return true, errors.New(assets.Texts["account_not_found"])
				}
				return true, err // какая-то ошибка, не дело не в блокировке endpoint
			} else {
				return isPrivate, nil // успех!
			}
		}
	}
}

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

func task(bot *tgbotapi.BotAPI, db map[int64]account, config Config) {
	log.Printf("CRON started")
	for chatId, storedAccounts := range db {
		for accountName, oldPrivateStatus := range storedAccounts {
			log.Printf("CRON updating %s", accountName)
			newPrivateStatus, err := getPrivateStatus(strings.ToLower(accountName))
			if err != nil { //любая ошибка при проверке статуса
				log.Printf("CRON ERROR updating %s", accountName)
				break
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
			time.Sleep(updateStep) // проверка следующего аккаунта через 30 секунд
		}
	}
}

func MainHandler(resp http.ResponseWriter, _ *http.Request) {
	resp.Write([]byte("<html><head><title>InstasnitchBot</title></head><body>Hi there! I'm InstasnitchBot!<br>I can do some shit.<br>You can get me at <a href=\"https://t.me/instasnitchbot\">https://t.me/instasnitchbot</a></body></html>"))
}

func main() {
	////http.HandleFunc("/", MainHandler)
	////go http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	// loading config
	config := getConfig()

	// setting up log
	f, err := os.OpenFile(config.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("ERROR opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

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

	//setting up cron
	s := gocron.NewScheduler(time.UTC)
	_, errS := s.Every(updatePeriod).Minutes().Do(task, bot, db, config)
	if errS != nil {
		log.Printf("CRON ERROR %v", errS)
	}
	s.StartAsync()

	for update := range updates {

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
			log.Printf("ADD try %s BY %s (ID %d)", update.Message.Text, update.Message.From.UserName, update.Message.From.ID)
			if _, ok := db[chatId]; !ok {
				db[chatId] = make(account)
			}

			if len(db[chatId]) > 2 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts["limit_of_accounts"])
				bot.Send(msg)
				continue
			}

			if strings.Contains(update.Message.Text, " ") {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
				bot.Send(msg)
				continue
			}

			privateStatus, err := getPrivateStatus(strings.ToLower(update.Message.Text))

			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], err.Error()))
				bot.Send(msg)
				continue
			}

			newAccountName := strings.ToLower(update.Message.Text)
			db[chatId][newAccountName] = privateStatus
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_added"], newAccountName))
			log.Printf("ADD success %s BY %s (ID %d)", newAccountName, update.Message.From.UserName, update.Message.From.ID)
			msg.ParseMode = "HTML"
			saveData(db, config)
			bot.Send(msg)
		}

	}
}
