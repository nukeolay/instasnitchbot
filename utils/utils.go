package utils

import (
	"encoding/json"
	"instasnitchbot/models"
	"io/ioutil"
	"log"
)

func SaveDb(db map[int64]models.Account, config models.Config) {
	file, err := json.MarshalIndent(db, "", " ")
	ioutil.WriteFile(config.DbName, file, 0644)
	if err != nil {
		log.Printf("SAVE DB ERROR %v", err)
	}
}

func LoadDb(config models.Config) (map[int64]models.Account){
	var db = map[int64]models.Account{}
	file, err := ioutil.ReadFile(config.DbName)
	if err != nil {
		log.Printf("LOAD DB ERROR %v", err)
	} else {
		db = map[int64]models.Account{}
		json.Unmarshal([]byte(file), &db)
	}
	return db
}
