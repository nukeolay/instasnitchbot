package main

import (
	"fmt"
	"instasnitchbot/api"
	"instasnitchbot/assets"
	"instasnitchbot/handlers"
	"instasnitchbot/utils"
	"log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// initializing
	loginCountdown := 0
	isTaskFinished := true
	config := utils.GetConfig()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	workingPath := path.Dir(filename)

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

	if insta == nil { // import error
		insta = api.GetNewApi(igAccounts)
		if insta == nil { // login error
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
	var updates tgbotapi.UpdatesChannel
	if config.UseWebhook {
		updates = bot.ListenForWebhook("/" + bot.Token)
	} else {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = bot.GetUpdatesChan(u)
	}

	db := utils.LoadDb(config)

	//setting up cron update account
	log.Println("----------------SETTING UP CRON----------------")
	cronStatusUpdater := gocron.NewScheduler(time.UTC)
	_, errCronStatusUpdater := cronStatusUpdater.Every(config.UpdateStatusPeriod).Minutes().Do(handlers.TaskStatusUpdater, bot, &insta, db, igAccounts, config, &loginCountdown, &isTaskFinished)
	if errCronStatusUpdater != nil {
		log.Printf("START CRON ERROR: %v", errCronStatusUpdater)
		handlers.SendAdmin(config.AdminChatId, bot, fmt.Sprintf("START CRON ERROR: %v", errCronStatusUpdater))
	} else {
		cronStatusUpdater.StartAsync()
	}

	//-----------------------------------HANDLING UPDATES-----------------------------------//
	for update := range updates {

		// if insta is nil
		if insta == nil {
			if update.Message == nil { // ignore all data except user messages
				continue
			}
			locale := db[update.Message.Chat.ID].Locale
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, assets.Texts[locale]["panic"])
			bot.Send(msg)
			continue
		}

		if update.CallbackQuery != nil { // button press handling
			handlers.CallBackHandler(bot, update, db, config)
			continue
		} else if update.Message == nil { // ignore all data except user messages
			continue
		} else if update.Message.Command() != "" { // messages handling
			handlers.CommandHandler(bot, update, db, config)
			continue
		} else { // adding new account
			handlers.MessageHandler(workingPath, bot, update, db, config, insta)
			continue
		}
	}
}
