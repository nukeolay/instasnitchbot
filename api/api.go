package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ahmdrz/goinsta/v2"
)

var LoginRequiredError = goinsta.ErrorN{"login_required", "fail", ""}
var UserNotFoundError = goinsta.ErrorN{"User not found", "fail", "user_not_found"}

func GetPrivateStatus(insta *goinsta.Instagram, username string) (isPrivate bool, err error) {
	igUser, err := insta.Profiles.ByName(username)
	if err != nil {
		return true, err
	} else {
		return igUser.IsPrivate, nil
	}
}

func LoadLogins() (igAccounts map[string]string) {
	file, err := ioutil.ReadFile(".igAccounts")
	if err != nil {
		log.Printf("IGLOGINS ERROR load %v", err)
		os.Exit(3)
		return nil
	} else {
		json.Unmarshal([]byte(file), &igAccounts)
		return igAccounts
	}
}



func GetSavedApi(igAccounts map[string]string) *goinsta.Instagram {
	insta, errLoad := goinsta.Import(".goinsta")
	if errLoad != nil { // если ошибка импорта
		log.Printf("INSTA ERROR import: %v", errLoad)
		return insta // возвращаю или ноль, или залогиненный
	}
	log.Print("INSTA import success")
	return insta // возвращаю залогиненный по импорту
}

func GetNewApi(igAccounts map[string]string) *goinsta.Instagram {
	insta := InstaLogin(igAccounts)
	if insta != nil {
		InstaExport(insta)
	}
	return insta
}

func InstaExport(insta *goinsta.Instagram) {
	errExport := insta.Export(".goinsta")
	if errExport != nil {
		log.Printf("INSTA export error: %v", errExport)
	} else {
		log.Printf("INSTA export success")
	}
}

func InstaLogin(igAccounts map[string]string) *goinsta.Instagram {
	var insta *goinsta.Instagram
	var errLogin error
	for igLogin, igPassword := range igAccounts {
		insta = goinsta.New(igLogin, igPassword)
		errLogin = insta.Login()
		if errLogin != nil {
			log.Printf("LOGIN ERROR with %s: %v ", igLogin, errLogin)
			continue
		}
		log.Printf("LOGIN success with %s", igLogin)
		return insta
	}
	log.Println("LOGIN ERROR all attempts fail")
	return nil
}


















//TODO скорее всего эту функцию нужно использовать только, когда возникает ошибка с входом
//TODO а при загрузке программы нужно использовать что-то попроще, где не происходит обязательного логина
//TODO типа такого

// func GetApi(igAccounts map[string]string) *goinsta.Instagram {
// 	insta, errLoad := goinsta.Import(".goinsta")
// 	if errLoad != nil { // если ошибка импорта
// 		log.Printf("INSTA ERROR import: %v", errLoad)
// 		insta = InstaLogin(igAccounts)
// 		if insta != nil {
// 			InstaExport(insta)
// 		}
// 		return insta // возвращаю или ноль, или залогиненный
// 	}
// 	//TODO возможно этот кусоу кода не нужно включать?
// 	//TODO если импорт удался, то можно отдавать залогиненный ранее файл
// 	//TODO а проверка на логин будет проводиться

// 	errLogin := insta.Login()

// 	if errLogin != nil { // если ошибка входа с импортированным файлом
// 		insta = InstaLogin(igAccounts)
// 		if insta != nil {
// 			InstaExport(insta)
// 		}
// 		return insta // возвращаю или ноль, или залогиненный
// 	}

// 	log.Print("LOGIN success with import")
// 	return insta // возвращаю залогиненный по импорту
// }



// func GetApi(username string, password string) (insta *goinsta.Instagram, err error) {
// 	insta, errLoad := goinsta.Import(".goinsta")
// 	if errLoad != nil {
// 		log.Printf("INSTA import error: %v", errLoad)
// 		insta = goinsta.New(username, password)
// 		errLogin := insta.Login()
// 		if errLogin != nil {
// 			log.Printf("INSTA login error: %v", errLogin)
// 			return nil, errors.New("login error")
// 		} else {
// 			log.Printf("INSTA login success")
// 			errExport := insta.Export(".goinsta")
// 			if errExport != nil {
// 				log.Printf("INSTA export error: %v", errLoad)
// 			}
// 			return insta, nil
// 		}
// 	}
// 	return insta, nil
// }
