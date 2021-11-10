package api

import (
	"fmt"
	"instasnitchbot/assets"
	"log"
	"os"
	"path"
	"time"

	"github.com/Davincible/goinsta"
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
		fileName := fmt.Sprintf("instasnitch_%d.jpg", time.Now().UnixNano())
		fullpath := workingDirectory + "/" + fileName
		errMediaDownload := item.Download(workingDirectory, fileName)
		if errMediaDownload != nil {
			log.Printf("MEDIA ERROR download: %v ", errMediaDownload)
			msg := tgbotapi.NewMessage(chatID, assets.Texts["ru"]["media_download_error"])
			bot.Send(msg)
		} else {
			photoToSend := tgbotapi.NewDocumentUpload(chatID, fullpath)
			bot.Send(photoToSend)
			errRemove := os.Remove(fullpath)
			if errRemove != nil {
				log.Printf("MEDIA ERROR can't remove %s: %v ", fullpath, errRemove)
			}
		}
	} else {
		msg := tgbotapi.NewMessage(chatID, assets.Texts["ru"]["media_not_a_photo"])
		bot.Send(msg)
	}

}

func DownloadMedia(mediaUrl string, workingDirectory string, insta *goinsta.Instagram, bot *tgbotapi.BotAPI, chatID int64) {
	shortCode := path.Base(path.Dir(mediaUrl))
	var media *goinsta.FeedMedia
	var errGetMedia error
	if len(shortCode) > 12 { //TODO в url сторис media id размещен не там, где у постаб а без "/"
		media, errGetMedia = insta.GetMedia(shortCode)
	} else {
		media, errGetMedia = insta.GetMedia(shortcodeToInstaID(shortCode))
	}
	if errGetMedia != nil {
		log.Printf("MEDIA ERROR: %v ", errGetMedia)
		msg := tgbotapi.NewMessage(chatID, assets.Texts["ru"]["media_download_error"])
		bot.Send(msg)
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
