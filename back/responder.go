package main

import (
	"encoding/json"
	"log"
	"os"
)

type ScheduleJSON struct {
	Id        string   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Workday   []string `json:"workday,omitempty"`
	Stoptime  string   `json:"stoptime,omitempty"`
	Starttime string   `json:"starttime,omitempty"`
	Exclude   string   `json:"exclude,omitempty"`
}

type SnapsotJSON struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Keepdays int    `json:"keepdays,omitempty"`
}

type SchedulesJSON map[string]ScheduleJSON
type SnapsotsJSON map[string]SnapsotJSON

type Responder interface {
	getJson() error
	updateJSON() error
}

func getJson() (string, SchedulesJSON) {
	var cj SchedulesJSON

	respFile, err := os.OpenFile(Config["json"], os.O_RDONLY, 0)
	if err != nil {
		log.Println("getJSON: error reading file: ", err)
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)

	err = json.Unmarshal(bytes, &cj)
	if err != nil {
		log.Println(err)
	}

	return string(bytes), cj
}

func updateJSON(cj SchedulesJSON) error {
	out, _ := json.Marshal(cj)
	custFile, err := os.OpenFile(Config["json"], os.O_WRONLY|os.O_CREATE, 0644)
	defer custFile.Close()
	custFile.Truncate(0)
	custFile.Seek(0, 0)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = custFile.WriteString(string(out))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
