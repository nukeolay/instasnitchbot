package utils

import (
	"encoding/json"
	"instasnitchbot/models"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

func GetConfig() models.Config {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := models.Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	return configuration
}

func GetAccountSettings() models.Config {
	file, _ := os.Open("acSettings.json")
	decoder := json.NewDecoder(file)
	configuration := models.Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	return configuration
}

func GetRandomUpdateNextAccount(defaultPeriod int) int {
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := 30
	return defaultPeriod + rand.Intn(max-min+1) + min
}

// func SaveDb(db map[int64]models.Account, config models.Config) {
// 	file, err := json.MarshalIndent(db, "", " ")
// 	ioutil.WriteFile(config.DbName, file, 0644)
// 	if err != nil {
// 		log.Printf("SAVE DB ERROR: %v", err)
// 	}
// }

func SaveDb(db map[int64]*models.Account, config models.Config) {
	file, err := json.MarshalIndent(db, "", " ")
	ioutil.WriteFile(config.DbName, file, 0644)
	if err != nil {
		log.Printf("SAVE DB ERROR: %v", err)
	}
}

// func LoadDb(config models.Config) map[int64]models.Account {
// 	var db = map[int64]models.Account{}
// 	file, err := ioutil.ReadFile(config.DbName)
// 	if err != nil {
// 		log.Printf("LOAD DB ERROR: %v", err)
// 	} else {
// 		json.Unmarshal([]byte(file), &db)
// 		log.Println("LOAD DB success")
// 	}
// 	return db
// }

func LoadDb(config models.Config) map[int64]*models.Account {
	var db = map[int64]*models.Account{}
	file, err := ioutil.ReadFile(config.DbName)
	if err != nil {
		log.Printf("LOAD DB ERROR: %v", err)
	} else {
		json.Unmarshal([]byte(file), &db)
		log.Println("LOAD DB success")
	}
	return db
}