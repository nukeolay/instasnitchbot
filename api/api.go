package api

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Davincible/goinsta"
	// "github.com/tcnksm/go-input"
)

//var LoginRequiredError = goinsta.ErrorN{"login_required", "fail", ""}
var UserNotFoundError = goinsta.ErrorN{"User not found", "", "fail", "user_not_found"}
var TooManyRequestsError = goinsta.ErrorN{"User not found", "", "fail", "user_not_found"}

func getUserFromSearchResult(username string, searchResult *goinsta.SearchResult) (*goinsta.User, error) {
	for _, igUser := range searchResult.Users {
		if igUser.Username == username {
			return igUser, nil
		}
	}
	return nil, UserNotFoundError
}

func GetPrivateStatus(insta *goinsta.Instagram, username string) (isPrivate bool, err error) {
	igUser, err := insta.Profiles.ByName(username)
	if err != nil {
		if typedErr, ok := err.(goinsta.ErrorN); ok {
			if typedErr.ErrorType == "user_not_found" {
				log.Printf("GETPRIVATE STATUS ERROR for user %s: %s", username, typedErr.ErrorType)
				return true, UserNotFoundError
			} else {
				log.Printf("GETPRIVATE STATUS ERROR for user %s: %s", username, err.Error()[0:10])
				return true, typedErr
			}
		} else {
			return true, err
		}
	}
	return igUser.IsPrivate, nil
}

func LoadLogins() (igAccounts map[string]string) {
	file, err := ioutil.ReadFile(".igAccounts")
	if err != nil {
		log.Printf("IGLOGINS ERROR load: %v", err)
		os.Exit(3)
		return nil
	} else {
		log.Print("IGLOGINS load success")
		json.Unmarshal([]byte(file), &igAccounts)
		return igAccounts
	}
}

func GetSavedApi(igAccounts map[string]string) *goinsta.Instagram {
	insta, errLoad := goinsta.Import(".goinsta")
	if errLoad != nil { // impoort error
		log.Printf("INSTA ERROR import: %v", errLoad)
		return insta // return nil or logged in user
	}
	log.Print("INSTA import success")
	return insta // return imported logged in user
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