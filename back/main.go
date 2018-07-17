package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/uzhinskiy/lib.go/pconf"
)

var (
	configfile string
	HTTPAddr   string
	err        error
	authttl    int
	Config     pconf.ConfigType
)

func init() {
	var addr, port string
	flag.StringVar(&addr, "bind", "", "Address to listen for HTTP requests on")
	flag.StringVar(&port, "port", "8080", "Port to listen for HTTP requests on")
	flag.StringVar(&configfile, "config", "main.cfg", "Read configuration from this file")
	flag.Parse()

	Config = make(pconf.ConfigType)
	err := Config.Parse(configfile)

	if err != nil {
		log.Fatal("Error: ", err)
	}

	log.Println("Read from config ", len(Config), " items:", Config)
	if Config["bind"] != "" {
		addr = Config["bind"]
	}
	if Config["port"] != "" {
		port = Config["port"]
	}
	HTTPAddr = addr + ":" + port
	fmt.Println(HTTPAddr)
}

func main() {
	logTo := os.Stderr
	if logTo, err = os.OpenFile(Config["log_file"], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		log.Fatal(err)
	}
	defer logTo.Close()
	log.SetOutput(logTo)

	http.HandleFunc("/", Index)
	http.HandleFunc("/admin", Admin)
	http.HandleFunc("/scheduler", Scheduler)
	http.HandleFunc("/snapshots", Snapshots)
	http.HandleFunc("/auth", Auth)
	http.HandleFunc("/list", List)
	http.HandleFunc("/info", Info)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/create", Update)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/dump/", Dump)
	http.HandleFunc("/aws/list", AwsList)
	log.Println("HTTP server listening on", HTTPAddr)

	err := http.ListenAndServe(HTTPAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
