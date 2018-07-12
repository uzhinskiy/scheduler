package main

import (
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

/* Реализация для Scheduler */

func readJson(fname string) []byte {
	json_f, err := os.OpenFile(fname, os.O_RDONLY, 0)
	if err != nil {
		log.Println("getJSON: error reading file: ", err)
	}
	/* считать содержимое файла */
	fi, err := json_f.Stat()
	var bytes = make([]byte, fi.Size())
	json_f.Read(bytes)

	/*
		err = json.Unmarshal(bytes, &sj)
			if err != nil {
				log.Println(err)
			}
	*/
	return bytes
}

func writeJSON(fname string, jb []byte) error {
	custFile, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE, 0644)
	defer custFile.Close()
	custFile.Truncate(0)
	custFile.Seek(0, 0)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = custFile.WriteString(string(jb))
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

/*
https://stackoverflow.com/questions/33928175/how-to-pass-different-types-to-a-function
https://stackoverflow.com/questions/40720039/how-to-make-possible-to-return-structs-of-different-types-from-one-function-with
https://stackoverflow.com/questions/35657362/how-to-return-dynamic-type-struct-in-golang
https://stackoverflow.com/questions/24911993/golang-use-one-value-in-conditional-from-function-returning-multiple-arguments
*/
