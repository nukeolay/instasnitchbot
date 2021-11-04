package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var header = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36"

type IgResponseA1 struct {
	Graphql struct {
		User struct {
			ID        string `json:"id"`
			IsPrivate bool   `json:"is_private"`
		} `json:"user"`
	} `json:"graphql"`
}

type IgResponseTopSearch struct {
	Users []struct {
		Position int `json:"position"`
		User     struct {
			Pk        string `json:"pk"`
			Username  string `json:"username"`
			IsPrivate bool   `json:"is_private"`
		} `json:"user,omitempty"`
	} `json:"users"`
}

func GetPrivateStatusA1(accountName string) (isPrivate bool, err error) {
	log.Printf("ENDPOINT a1 used for fetching %s", accountName)
	url := fmt.Sprintf("https://www.instagram.com/%s/?__a=1", accountName)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", header)
	resp, errHttp := client.Do(req)
	if errHttp != nil {
		return true, NewEndpointErrorHttp(errHttp.Error())
	}
	var igResponse IgResponseA1
	errParsing := json.NewDecoder(resp.Body).Decode(&igResponse)
	if errParsing != nil {
		return true, NewEndpointErrorParsing(errParsing.Error())
	}
	if igResponse.Graphql.User.ID == "" {
		return true, NewEndpointErrorAccountNotFound(accountName)
	}
	return igResponse.Graphql.User.IsPrivate, nil
}

func GetPrivateStatusA1Channel(accountName string) (isPrivate bool, err error) {
	log.Printf("ENDPOINT a1/channel used for fetching %s", accountName)
	url := fmt.Sprintf("https://www.instagram.com/%s/channel/?__a=1", accountName)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", header)
	resp, errHttp := client.Do(req)
	if errHttp != nil {
		return true, NewEndpointErrorHttp(errHttp.Error())
	}
	var igResponse IgResponseA1
	errParsing := json.NewDecoder(resp.Body).Decode(&igResponse)
	if errParsing != nil {
		return true, NewEndpointErrorParsing(errParsing.Error())
	}
	if igResponse.Graphql.User.ID == "" {
		return true, NewEndpointErrorAccountNotFound(accountName)
	}
	return igResponse.Graphql.User.IsPrivate, nil
}

func GetPrivateStatusTopSearch(accountName string) (isPrivate bool, err error) {
	log.Printf("ENDPOINT topsearch used for fetching %s", accountName)
	url := fmt.Sprintf("https://www.instagram.com/web/search/topsearch/?query=%s", accountName)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", header)
	resp, errHttp := client.Do(req)
	if errHttp != nil {
		return true, NewEndpointErrorHttp(errHttp.Error())
	}
	var igResponseTopSearch IgResponseTopSearch
	errParsing := json.NewDecoder(resp.Body).Decode(&igResponseTopSearch)
	if errParsing != nil {
		return true, NewEndpointErrorParsing(errParsing.Error())
	}
	for _, igUser := range igResponseTopSearch.Users {
		if igUser.User.Username == accountName {
			return igUser.User.IsPrivate, nil
		}
	}
	return true, NewEndpointErrorAccountNotFound(accountName)
}
