package api

import (
	"fmt"
	"instasnitchbot/assets"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Davincible/goinsta"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func downloadFile(item goinsta.Item, workingDirectory string, bot *tgbotapi.BotAPI, chatID int64, locale string) {
	var fileName string
	if item.Videos != nil {
		fileName = fmt.Sprintf("instasnitch_%d.mp4", time.Now().UnixNano())
	} else {
		fileName = fmt.Sprintf("instasnitch_%d.jpg", time.Now().UnixNano())
	}
	fullpath := workingDirectory + "/" + fileName
	errMediaDownload := item.Download(workingDirectory, fileName)
	if errMediaDownload != nil {
		log.Printf("MEDIA ERROR download: %v", errMediaDownload)
		msg := tgbotapi.NewMessage(chatID, assets.Texts[locale]["media_download_error"])
		bot.Send(msg)
	} else {
		if item.Videos != nil {
			videoToSend := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(fullpath))
			_, uploadErr := bot.Send(videoToSend)
			if uploadErr != nil {
				log.Printf("MEDIA ERROR download: %v", uploadErr)
				msg := tgbotapi.NewMessage(chatID, assets.Texts[locale]["media_too_large_error"])
				bot.Send(msg)
				time.Sleep(time.Duration(120 * 1000000000)) // пауза перед запуском таска
			}
		} else {
			photoToSend := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(fullpath))
			bot.Send(photoToSend)
		}
		errRemove := os.Remove(fullpath)
		if errRemove != nil {
			log.Printf("MEDIA ERROR can't remove %s: %v ", fullpath, errRemove)
		}
	}
}

func DownloadMedia(mediaUrl string, workingDirectory string, insta *goinsta.Instagram, bot *tgbotapi.BotAPI, chatID int64, locale string) {
	var shortCode string
	var media *goinsta.FeedMedia
	var errGetMedia error

	if strings.Contains(mediaUrl, "stories") { //TODO проверить сторис
		shortCode = path.Base(mediaUrl)
		shortCode = strings.Split(shortCode, "?")[0]
		media, errGetMedia = insta.GetMedia(shortCode)
	} else {
		shortCode = path.Base(path.Dir(mediaUrl))
		media, errGetMedia = insta.GetMedia(shortcodeToInstaID(shortCode))
	}

	if errGetMedia != nil {
		log.Printf("MEDIA ERROR can't get insta.GetMedia for url: %s: %s", mediaUrl, errGetMedia.Error()[0:10])
		msg := tgbotapi.NewMessage(chatID, assets.Texts[locale]["media_download_error"])
		bot.Send(msg)
		return
	}
	log.Printf("MEDIA get success: %s ", mediaUrl)
	for _, item := range media.Items {
		if item.CarouselMedia != nil {
			for _, carItem := range item.CarouselMedia {
				downloadFile(carItem, workingDirectory, bot, chatID, locale)
			}
		} else {
			downloadFile(item, workingDirectory, bot, chatID, locale)
		}
	}
}
