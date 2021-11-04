package api

import (
	"errors"
	"log"

	"github.com/ahmdrz/goinsta/v2"
)

func GetPrivateStatus(insta *goinsta.Instagram, username string) (isPrivate bool, err error) {
	igUser, err := insta.Profiles.ByName(username)
	if err != nil {
		return true, err
	} else {
		return igUser.IsPrivate, nil
	}
}

func GetApi(username string, password string) (insta *goinsta.Instagram, err error) {
	insta, errLoad := goinsta.Import(".goinsta")
	if errLoad != nil {
		log.Printf("INSTA import error: %v", errLoad)
		insta = goinsta.New(username, password)
		errLogin := insta.Login()
		if errLogin != nil {
			log.Printf("INSTA login error: %v", errLogin)
			return nil, errors.New("login error")
		} else {
			log.Printf("INSTA login success")
			errExport := insta.Export(".goinsta")
			if errExport != nil {
				log.Printf("INSTA export error: %v", errLoad)
			}
			return insta, nil
		}
	}
	return insta, nil
}