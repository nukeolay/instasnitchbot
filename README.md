InstasnitchBot

*USE IT AT YOUR OWN RISK*

Here are some steps you have to do before running bot for the first time
1. In bot folder create ".igAccounts" file with following structure (or edit if it exists)
    {
    "Instagram_account_name": "Instagram_account_password"
    }
    - enter your instagram account credentials
    - your password is only required to get access to Instagram API and not send to anyone
    - it is recommended to create new instagram account to use woth this bot
    - creator of this bot is not responsible for any problems that can occure when youre using this bot

2. In bot folder create ".config" file with following structure (or edit if it exists)
   {
    "TelegramBotToken": "_____", // you can get token from @BotFather
   	"LogFileName": ".log",
   	"DbName": ".db",
   	"UseWebhook": false, // not tested yet
   	"TryLoginPeriod": 5, // number of UpdateStatusPeriod (5*10)
   	"UpdateStatusPeriod": 15, // minutes
   	"UpdateNextAccount": 90, // seconds
   	"SnitchLimit": 3,
   	"Port": "80",
   	"AdminChatId": _____ // it is used to send you information from the bot when new users connect
   }

    To get AdminChatId (actually your ChatId) you should do followong steps when first time starting bot:
   - start "instasnitchbot.exe
   - send in Telegram "/start" command to your bot
   - wait for 1 minute
   - close running "instasnitchbot.exe"
   - open ".db" file in bot folder and look for number (for example 1234567890), it should be in the second line of ".db" file
    - add this number to ".config" file, it should look like "AdminChatId": 1234567890
    - save ".db" file
    - start bot again
