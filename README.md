# InstasnitchBot
#
# in root folder of the project create "config.json" 
# with following structure
#   {
#   	"TelegramBotToken": "_____",
#   	"LogFileName": ".log",
#   	"DbName": ".db",
#   	"UseWebhook": false,
#   	"TryLoginPeriod": 5, // number of UpdateStatusPeriod (5*10)
#   	"UpdateStatusPeriod": 10, // minutes
#   	"UpdateNextAccount": 30, // seconds
#   	"SnitchLimit": 3,
#   	"Port": "80",
#   	"AdminChatId": _____
#   }