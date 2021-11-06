package models

type Config struct {
	TelegramBotToken   string
	LogFileName        string
	DbName             string
	UseWebhook         bool
	UpdateStatusPeriod int
	UpdateNextAccount  int
	SnitchLimit        int
	Port               string
	AdminChatId        int64
}

type Account map[string]bool