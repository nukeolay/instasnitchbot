package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
	IgUsername1        string
	IgPassword1        string
	IgUsername2        string
	IgPassword2        string
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

func getPrivateStatus(insta *goinsta.Instagram, username string) (isPrivate bool, err error) {
	igUser, err := insta.Profiles.ByName(username)
	if err != nil {
		return true, err
	} else {
		return igUser.IsPrivate, nil
	}
}

func getApi(username string, password string) (insta *goinsta.Instagram, err error) {
	insta, errLoad := goinsta.Import(".goinsta")
	if errLoad != nil {
		log.Printf("INSTA import error: %v", errLoad)
		insta = goinsta.New(username, password)
		errLogin := insta.Login()
		if errLogin != nil {
			log.Printf("INSTA login error: %v", errLogin)
			return nil, errors.New("login error")
		} else {
			log.Printf("INSTA login success")
			errExport := insta.Export(".goinsta")
			if errExport != nil {
				log.Printf("INSTA export error: %v", errLoad)
			}
			return insta, nil
		}
	}
	return insta, nil
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

func task(bot *tgbotapi.BotAPI, insta *goinsta.Instagram, db map[int64]account, config Config) {
	log.Printf("CRON started")
	for chatId, storedAccounts := range db {
		for accountName, oldPrivateStatus := range storedAccounts {
			log.Printf("CRON updating %s", accountName)
			newPrivateStatus, err := getPrivateStatus(insta, strings.ToLower(accountName))
			if err != nil { // любая ошибка при проверке статуса
				log.Printf("CRON ERROR updating %s, %v", accountName, err)
				insta, err = getApi(config.IgUsername1, config.IgPassword1)
				if err != nil { // ошибка авторизации
					log.Printf("CRON ERROR login %s, %v", config.IgUsername1, err)
					insta, err = getApi(config.IgUsername2, config.IgPassword2)
					if err != nil { // ошибка авторизации
						log.Printf("CRON ERROR login %s, %v", config.IgUsername2, err)
						break
					} else {
						newPrivateStatus, err = getPrivateStatus(insta, strings.ToLower(accountName))
						if err != nil { // это ошибка не связанная с логином, возможно поменялось имя акканта, пропустить обновление
							log.Printf("CRON ERROR updating %s, %v", accountName, err)
							continue
						}
					}
				} else {
					newPrivateStatus, err = getPrivateStatus(insta, strings.ToLower(accountName))
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
	////http.HandleFunc("/", MainHandler)
	////go http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	config := getConfig()

	// setting up log
	f, err := os.OpenFile(config.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("ERROR opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// instagram login
	insta, err := getApi(config.IgUsername1, config.IgPassword1)
	if err != nil {
		log.Printf("INSTA error getApi: %v", err)
		insta, err = getApi(config.IgUsername2, config.IgPassword2)
		if err != nil {
			log.Panic(err)
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
	_, errS := s.Every(config.UpdateStatusPeriod).Minutes().Do(task, bot, insta, db, config)
	if errS != nil {
		log.Printf("CRON ERROR update status %v", errS)
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
			if update.Message.From.IsBot || update.Message.ForwardFrom.IsBot {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["do_not_work_with_bots"]))
				bot.Send(msg)
				continue
			}

			privateStatus, err := getPrivateStatus(insta, strings.ToLower(update.Message.Text))
			if err != nil { // любая ошибка при проверке статуса
				log.Printf("ADD error %s, %v", strings.ToLower(update.Message.Text), err)
				insta, err = getApi(config.IgUsername1, config.IgPassword1)
				if err != nil { // ошибка авторизации
					log.Printf("ADD ERROR login %s, %v", config.IgUsername1, err)
					insta, err = getApi(config.IgUsername2, config.IgPassword2)
					if err != nil { // ошибка авторизации
						log.Printf("ADD ERROR login %s, %v", config.IgUsername2, err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], err.Error()))
						bot.Send(msg)
						continue
					} else {
						privateStatus, err = getPrivateStatus(insta, strings.ToLower(update.Message.Text))
						if err != nil { // это ошибка не связанная с логином, возможно поменялось имя акканта
							log.Printf("ADD error %s, %v", strings.ToLower(update.Message.Text), err)
							msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
							bot.Send(msg)
							continue
						}
					}
				} else {
					privateStatus, err = getPrivateStatus(insta, strings.ToLower(update.Message.Text))
					if err != nil { // это ошибка не связанная с логином, возможно поменялось имя акканта
						log.Printf("ADD error %s, %v", strings.ToLower(update.Message.Text), err)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(assets.Texts["account_add_error"], assets.Texts["account_not_found"]))
						bot.Send(msg)
						continue
					}
				}
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
