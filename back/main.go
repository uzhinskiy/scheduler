package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"./conf"
	"github.com/uzhinskiy/htpass"
)

var (
	configfile string
	HTTPAddr   string
	err        error
	authttl    int
)

type InstanceJSON struct {
	Id        string   `json:"id,omitempty"`
	Name      string   `json:"name,omitempty"`
	Workday   []string `json:"workday,omitempty"`
	Stoptime  string   `json:"stoptime,omitempty"`
	Starttime string   `json:"starttime,omitempty"`
}

type InstancesJSON map[string]InstanceJSON

func init() {
	var addr, port string
	flag.StringVar(&addr, "bind", "", "Address to listen for HTTP requests on")
	flag.StringVar(&port, "port", "8080", "Port to listen for HTTP requests on")
	flag.StringVar(&configfile, "config", "main.cfg", "Read configuration from this file")
	flag.Parse()

	conf.Config = make(conf.ConfigType)
	err := conf.Config.Parse(configfile)

	checkError(err, 1)

	log.Println("Read from config ", len(conf.Config), " items")
	if conf.Config["bind"] != "" {
		addr = conf.Config["bind"]
	}
	if conf.Config["port"] != "" {
		port = conf.Config["port"]
	}
	HTTPAddr = addr + ":" + port
	fmt.Println(HTTPAddr)
}

func checkError(err error, fatal int) {
	if err != nil {
		if fatal == 1 {
			log.Fatal("Error: ", err)
		} else {
			log.Println("Error: ", err)
		}
	}
}

func getJson() (string, InstancesJSON) {
	var cj InstancesJSON

	respFile, err := os.OpenFile(conf.Config["json"], os.O_RDONLY, 0)
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

func updateJSON(cj InstancesJSON) error {
	out, _ := json.Marshal(cj)
	custFile, err := os.OpenFile(conf.Config["json"], os.O_WRONLY|os.O_CREATE, 0644)
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

func main() {
	logTo := os.Stderr

	if logTo, err = os.OpenFile(conf.Config["log_file"], os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600); err != nil {
		log.Fatal(err)
	}
	defer logTo.Close()
	log.SetOutput(logTo)

	http.HandleFunc("/", Index)
	http.HandleFunc("/admin", Admin)
	http.HandleFunc("/auth", Auth)
	http.HandleFunc("/list", List)
	http.HandleFunc("/info", Info)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/create", Update)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/dump/", Dump)
	log.Println("HTTP server listening on", HTTPAddr)

	err := http.ListenAndServe(HTTPAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

/* функция для обработки подключившихся клиентов */
func Index(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	base := conf.Config["document_root"]
	/* если отсутствует запрос к конкретному файлу – показать индексный файл */
	if file == "/" {
		file = "/index.html"
	}

	code := http.StatusOK
	/* если не удалось загрузить нужный файл – показать сообщение о 404-ой ошибке */
	respFile, err := os.OpenFile(base+file, os.O_RDONLY, 0)
	if err != nil {
		log.Println(err)
		file = "/404.html"
		respFile, err = os.OpenFile(base+file, os.O_RDONLY, 0)
		code = http.StatusNotFound
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	contentType := mime.TypeByExtension(path.Ext(file))
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)
	log.Println(r.Method, "\t", r.URL.Path, "\t", code, "\t", r.UserAgent())
	/* отправить его клиенту */
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", conf.Config["version"])
	w.Write(bytes)
}

func List(w http.ResponseWriter, r *http.Request) {
	custJSONs, _ := getJson()
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Server", conf.Config["version"])
	log.Println(r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
	fmt.Fprint(w, fmt.Sprintf("%s", custJSONs))
}

func Info(w http.ResponseWriter, r *http.Request) {
	_, cj := getJson()
	queryValues := r.URL.Query()
	id := queryValues.Get("id")
	j, err := json.Marshal(cj[id])
	if err != nil {
		log.Println(err)
	}
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Server", conf.Config["version"])
	log.Println(r.Method, "\t", r.URL.RequestURI(), "\t", http.StatusOK, "\t", r.UserAgent())
	fmt.Fprint(w, fmt.Sprintf("%s", j))
}

func Update(w http.ResponseWriter, r *http.Request) {
	var new_cj InstanceJSON
	var err error

	_, cj := getJson()

	r.ParseForm()
	queryValues := r.PostFormValue
	id := queryValues("id")
	new_cj.Id = queryValues("id")
	new_cj.Name = queryValues("name")
	new_cj.Starttime = queryValues("starttime")
	new_cj.Stoptime = queryValues("stoptime")
	new_cj.Workday = r.Form["wd"]

	if new_cj.Id != "" && new_cj.Name != "" && new_cj.Starttime != "" && new_cj.Stoptime != "" {
		cj[id] = new_cj
		err = updateJSON(cj)
	} else {
		err = fmt.Errorf("error")
	}

	if err != nil {
		log.Println(err)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Server", conf.Config["version"])
		log.Println(r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", r.UserAgent())
		fmt.Fprint(w, "<h1>Error while file saving</h1><a href='/'>back</a>")
	} else {
		log.Println(r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	_, cj := getJson()
	queryValues := r.PostFormValue
	id := queryValues("id")
	_, err := json.Marshal(cj[id])
	if err != nil {
		log.Println(err)
	}

	delete(cj, id)
	err = updateJSON(cj)
	if err != nil {
		log.Println(err)
	}

	if err != nil {
		log.Println(err)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Server", conf.Config["version"])
		log.Println(r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", r.UserAgent())
		fmt.Fprint(w, "<h1>Error while file saving</h1><a href='/'>back</a>")
	} else {
		log.Println(r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	}
}

func Dump(w http.ResponseWriter, r *http.Request) {
	//queryValues := r.URL.Query()
	urlPart := strings.Split(r.URL.Path, "/")
	dumpWhat := urlPart[2]
	log.Println(r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Server", conf.Config["version"])

	switch dumpWhat {
	case "config":
		{
			fmt.Fprint(w, "<pre><ul>")
			for k, cfg := range conf.Config {
				fmt.Fprintf(w, "<li>%s = %s</li>", k, cfg)
			}
			fmt.Fprint(w, "</ul></pre>")
		}
	case "log":
		{
			respFile, err := os.OpenFile(conf.Config["log_file"], os.O_RDONLY, 0)
			if err != nil {
				log.Println(err)
			}
			fi, err := respFile.Stat()
			if err != nil {
				log.Println(err)
			}
			var bytes = make([]byte, fi.Size())
			respFile.Read(bytes)

			fmt.Fprintf(w, "<pre>%s</pre>", bytes)
		}
	}
}

func Admin(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	base := conf.Config["document_root"]

	auth_cookie, _ := r.Cookie("isauth")
	/* если отсутствует запрос к конкретному файлу – показать индексный файл */
	if auth_cookie == nil || auth_cookie.Value == "no" {
		file = "/login.html"
	} else {
		file = "/admin.html"
	}
	code := http.StatusOK
	/* если не удалось загрузить нужный файл – показать сообщение о 404-ой ошибке */
	respFile, err := os.OpenFile(base+file, os.O_RDONLY, 0)
	if err != nil {
		log.Println(err)
		file = "/404.html"
		respFile, err = os.OpenFile(base+file, os.O_RDONLY, 0)
		code = http.StatusNotFound
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	contentType := mime.TypeByExtension(path.Ext(file))
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)
	log.Println(r.Method, "\t", r.URL.Path, "\t", code, "\t", r.UserAgent())
	/* отправить его клиенту */
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", conf.Config["version"])
	w.Write(bytes)
}

func Auth(w http.ResponseWriter, r *http.Request) {
	htp := make(htpass.HTPassFile)
	err := htp.OpenHTPASSFile(conf.Config["pswdfile"])
	queryValues := r.PostFormValue
	res, err := htp.Auth(queryValues("username"), queryValues("passwd"))
	log.Printf("%v\t%v\t%v\n", htpass.IsAuth, res, err)
	if err != nil {
		fmt.Println(err)
	}
	authttl, _ = strconv.Atoi(conf.Config["authttl"])
	if res {
		expiration := time.Now().Add(time.Duration(authttl) * time.Minute)
		cookie := http.Cookie{Name: "isauth", Value: "yes", Expires: expiration}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/admin", http.StatusMovedPermanently)
	} else {
		expiration := time.Now().Add(time.Duration(authttl) * time.Minute)
		cookie := http.Cookie{Name: "isauth", Value: "no", Expires: expiration}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/login.html", http.StatusMovedPermanently)
	}
}
