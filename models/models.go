package models

type Config struct {
	TelegramBotToken   string
	LogFileName        string
	DbName             string
	UseWebhook         bool
	TryLoginPeriod     int
	UpdateStatusPeriod int
	UpdateNextAccount  int
	SnitchLimit        int
	Port               string
	AdminChatId        int64
}

type Account struct {
	Locale     string
	IgAccounts map[string]bool
}