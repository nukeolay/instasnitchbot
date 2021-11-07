package api

import (
	"instasnitchbot/assets"
	"log"
	"os"
	"path"

	"github.com/ahmdrz/goinsta/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func getCharNumber(inputChar string) int64 {
	for pos, char := range "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_" {
		if inputChar == string(char) {
			return int64(pos)
		}
	}
	return 0
}

func shortcodeToInstaID(shortcode string) int64 {
	var id int64 = 0
	for _, char := range shortcode {
		id = (id * 64) + getCharNumber(string(char))
	}
	return id
}

func downloadPhoto(item goinsta.Item, workingDirectory string, bot *tgbotapi.BotAPI, chatID int64) {
	if item.Videos == nil {
		imgs, _, errMediaDownload := item.Download(workingDirectory, "")
		if errMediaDownload != nil {
			log.Printf("MEDIA ERROR download: %v ", errMediaDownload)
			msg := tgbotapi.NewMessage(chatID, assets.Texts["media_download_error"])
			bot.Send(msg)
		} else {
			photoToSend := tgbotapi.NewDocumentUpload(chatID, imgs)
			bot.Send(photoToSend)
			errRemove := os.Remove(imgs)
			if errRemove != nil {
				log.Printf("MEDIA ERROR can't remove %s: %v ", imgs, errRemove)
			}
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, assets.Texts["media_not_a_photo"])
		bot.Send(msg)
	}

}

func DownloadMedia(mediaUrl string, workingDirectory string, insta *goinsta.Instagram, bot *tgbotapi.BotAPI, chatID int64) {
	shortCode := path.Base(path.Dir(mediaUrl))
	media, errGetMedia := insta.GetMedia(shortcodeToInstaID(shortCode))
	if errGetMedia != nil {
		log.Printf("MEDIA ERROR: %v ", errGetMedia)
		return
	}
	log.Printf("MEDIA get success: %s ", shortCode)
	for _, item := range media.Items {
		if item.CarouselMedia != nil {
			for _, carItem := range item.CarouselMedia {
				downloadPhoto(carItem, workingDirectory, bot, chatID)
			}
		} else {
			downloadPhoto(item, workingDirectory, bot, chatID)
		}
	}
}
