# InstasnitchBot
#
# in bot folder create ".config" file
# with following structure
#   {
#   	"TelegramBotToken": "_____", // you can get token from @BotFather
#   	"LogFileName": ".log",
#   	"DbName": ".db",
#   	"UseWebhook": false, // not tested yet
#   	"TryLoginPeriod": 5, // number of UpdateStatusPeriod (5*10)
#   	"UpdateStatusPeriod": 15, // minutes
#   	"UpdateNextAccount": 90, // seconds
#   	"SnitchLimit": 3,
#   	"Port": "80",
#   	"AdminChatId": _____ // it is used to send you information from the bot if new users connect
#   }
#
# To get AdminChatId (actually your ChatId) you should do followong steps when first time starmig bot:
# 1. start "instasnitchbot.exe
# 2. send in Telegram "/start" command to your bot
# 3. wait for 1 minute
# 4. close running "instasnitchbot.exe"
# 5. open ".db" file in bot folder and look for number (for example 1234567890), it should be in the second line of ".db" file
# 6. add this number to ".config" file, it should look like "AdminChatId": 1234567890
# 7. save ".db" file
# 8. start bot again