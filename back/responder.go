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

type SnapshotJSON struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Keepdays int    `json:"keepdays,omitempty"`
}

type SchedulesJSON map[string]ScheduleJSON
type SnapshotsJSON map[string]SnapshotJSON

type Responder interface {
	getJson()
	updateJSON()
}

/* Реализация для Scheduler */

func (sj *SchedulesJSON) getJson() string {
	respFile, err := os.OpenFile(Config["scheduler"], os.O_RDONLY, 0)
	if err != nil {
		log.Println("getJSON: error reading file: ", err)
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)

	err = json.Unmarshal(bytes, &sj)
	if err != nil {
		log.Println(err)
	}

	return string(bytes)
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

/* Реализация для Snapshots */

func (sj *SnapshotsJSON) getJson() string {
	respFile, err := os.OpenFile(Config["snapshots"], os.O_RDONLY, 0)
	if err != nil {
		log.Println("getJSON: error reading file: ", err)
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)

	err = json.Unmarshal(bytes, &sj)
	if err != nil {
		log.Println(err)
	}

	return string(bytes)
}

/*
https://stackoverflow.com/questions/33928175/how-to-pass-different-types-to-a-function
https://stackoverflow.com/questions/40720039/how-to-make-possible-to-return-structs-of-different-types-from-one-function-with
https://stackoverflow.com/questions/35657362/how-to-return-dynamic-type-struct-in-golang
https://stackoverflow.com/questions/24911993/golang-use-one-value-in-conditional-from-function-returning-multiple-arguments
*/
