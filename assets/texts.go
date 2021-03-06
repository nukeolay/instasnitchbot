package assets

var Texts = map[string]map[string]string{
	"ru": {
		"choose_language":           "Выберите язык:",
		"instructions":              "На связи <b><u>Instasnitch</u></b> 🤖\n\n<b><u>Я умею:</u></b>\n\n✅ наблюдать 👀 за закрытыми аккаунтами в Instagram и уведомлять тебя, когда они становятся открытыми.\n\n✅ сохранять и присылать изображения и видео, размещенные в Instagram.\n\n<b><u>Почему это работает:</u></b>\n\nВсе владельцы закрытых аккаунтов рано или поздно открывают их на непродолжительное время, например, для участия в “гивах”. Я помогу тебе не упустить этот момент. Всё происходит анонимно: мне не нужен ни твой профиль Instagram, ни доступ к нему.\n\n<b><u>Как это работает:</u></b>\n\n🔴 Чтобы я начал наблюдать за аккаунтом, напиши мне его имя.\n\n🔴 Чтобы я прислал тебе изображение или видео из Instagram, просто отправь мне ссылку на пост, историю или reels.\n\nЕсли хочешь связаться с моим создателем, пиши @fuckyouryankeebluejeans",
		"endpoint_error":            "сервер перегружен, нужно попробовать позже",
		"account_deleted":           "🙈 перестал наблюдать за <u>%s</u>",
		"account_choose_to_delete":  "🗑️ выберите аккаунт, за которым больше не требуется наблюдение",
		"account_added":             "🧐 начал наблюдать за <u>%s</u>",
		"account_added_not_private": "🧐 начал наблюдать за <u>%s</u>, но он и так открытый 🟢. Я точно должен за ним наблюдать? Если нет, удали его (/del) чтобы я не тратил свои ресурсы",
		"account_not_found":         "аккаунт не найден",
		"account_add_error":         "⛔ ошибка при добавлении аккаунта (%s)",
		"account_list_is_empty_now": "🙄 больше я ни за кем не наблюдаю",
		"account_list_is_empty":     "🙄 я ни за кем не наблюдаю",
		"account_is_not_private":    "🟢 псс, аккаунт <u>%s</u> перестал быть приватным 😉",
		"account_is_private":        "🔴 ну вот, аккаунт <u>%s</u> стал приватным ☹️",
		"unknown_command":           "🤷‍♀️ я не знаю такой команды",
		"limit_of_accounts":         "⛔ одновременно я могу наблюдать только за %d аккаунтами.\n\nТы можешь удалить один из добавленных аккаунтов (/del), а затем добавить новый",
		"do_not_work_with_bots":     "⛔ я с ботами не работаю",
		"media_not_a_photo":         "⛔ ой, это было видео, я их не загружаю",
		"media_download_error":      "⛔ что-то пошло не так (убедись, что контент размещен в открытом аккаунте Instagram)",
		"media_too_large_error":     "⛔ файл большой, а я маленький, не получилось его загрузить, извини",
		"panic":                     "💩 я сломался, попробуй написать мне позже",
		"button_accounts":           "👥 Показать список аккаунтов",
		"button_delete":             "🗑️ Перестать следить за аккаунтом",
		"button_help":               "❔ Помощь",
	},
	"en": {
		"choose_language":           "Choose your language:",
		"instructions":              "Sup! I`m <b><u>Instasnitch</u></b> 🤖\n\n<b><u>I can:</u></b>\n\n✅ track 👀 for private Instagram accounts and inform you when they become public.\n\n✅ save and send you images and videos from Instagram posts, reels and stories.\n\n<b><u>Why it works:</u></b>\n\nAll private Instagram account owners sooner or later makes them public for a short time occasionally. I help you not to miss this moment. Everything happens anonymously: I don't need your Instagram account or access to it.\n\n<b><u>How it works:</u></b>\n\n🔴 To start track private account, just send me its name.\n\n🔴 To get images or videos from Instagram, just send me the link to post, stories or reels.\n\nFeel free to contact my creator @fuckyouryankeebluejeans",
		"endpoint_error":            "sorry, the server is busy, please try again later",
		"account_deleted":           "🙈 I`m not tracking for <u>%s</u> anymore",
		"account_choose_to_delete":  "🗑️ choose the account you no longer want to track",
		"account_added":             "🧐 start tracking for <u>%s</u>",
		"account_added_not_private": "🧐 start tracking for <u>%s</u>, but this account already not private 🟢. Am I really supposed to be tracking him? If not, please delete it (/ del) so I don't waste my resourses on it",
		"account_not_found":         "account not found",
		"account_add_error":         "⛔ can't add account (%s)",
		"account_list_is_empty_now": "🙄 I'm not tracking for anyone anymore",
		"account_list_is_empty":     "🙄 I'm not tracking for anyone",
		"account_is_not_private":    "🟢 good news! <u>%s</u> is not private anymore 😉",
		"account_is_private":        "🔴 sad news. <u>%s</u> is private now ☹️",
		"unknown_command":           "🤷‍♀️ I don't know this command",
		"limit_of_accounts":         "⛔ even I have limits, I can track for only %d accounts at the same time.\n\nYou can delete one of the added accounts (/ del) and then add a new one.",
		"do_not_work_with_bots":     "⛔ I don't work with bots",
		"media_not_a_photo":         "⛔ oops, it was a video, I don't download videos",
		"media_download_error":      "⛔ something went wrong (make sure it is a proper Instagram link and it is posted on a public account)",
		"media_too_large_error":     "⛔ this file is too big, and I am too small, I can't download it, sorry",
		"panic":                     "💩 I'm not feeling well, please try again later",
		"button_accounts":           "👥 Show accounts",
		"button_delete":             "🗑️ Stop tracking account",
		"button_help":               "❔ Help",
	},
}
