package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/uzhinskiy/lib.go/helpers"
	"github.com/uzhinskiy/lib.go/htpass"
)

/* функция для обработки подключившихся клиентов */
func Index(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	base := Config["document_root"]
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
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", code, "\t", r.UserAgent())
	/* отправить его клиенту */
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", Config["version"])
	w.Write(bytes)
}
func Scheduler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	base := Config["document_root"]

	auth_cookie, _ := r.Cookie("isauth")
	/* если отсутствует запрос к конкретному файлу – показать индексный файл */
	if auth_cookie == nil || auth_cookie.Value == "no" {
		file = "/login.html"
	} else {
		file = "/sched.html"
	}
	code := http.StatusOK
	/* если не удалось загрузить нужный файл – показать сообщение о 404-ой ошибке */
	respFile, err := os.OpenFile(base+file, os.O_RDONLY, 0)
	if err != nil {
		log.Println(r.RemoteAddr, "\t", err)
		file = "/404.html"
		respFile, err = os.OpenFile(base+file, os.O_RDONLY, 0)
		code = http.StatusNotFound
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	contentType := mime.TypeByExtension(path.Ext(file))
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", code, "\t", r.UserAgent())
	/* отправить его клиенту */
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", Config["version"])
	w.Write(bytes)
}

func Snapshots(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path
	base := Config["document_root"]

	auth_cookie, _ := r.Cookie("isauth")
	/* если отсутствует запрос к конкретному файлу – показать индексный файл */
	if auth_cookie == nil || auth_cookie.Value == "no" {
		file = "/login.html"
	} else {
		file = "/snap.html"
	}
	code := http.StatusOK
	/* если не удалось загрузить нужный файл – показать сообщение о 404-ой ошибке */
	respFile, err := os.OpenFile(base+file, os.O_RDONLY, 0)
	if err != nil {
		log.Println(r.RemoteAddr, "\t", err)
		file = "/404.html"
		respFile, err = os.OpenFile(base+file, os.O_RDONLY, 0)
		code = http.StatusNotFound
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	contentType := mime.TypeByExtension(path.Ext(file))
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", code, "\t", r.UserAgent())
	/* отправить его клиенту */
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", Config["version"])
	w.Write(bytes)
}

func List(w http.ResponseWriter, r *http.Request) {
	var custJSONs []byte
	obj := r.URL.Query().Get("object")
	custJSONs = readJson(Config[obj])

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Server", Config["version"])
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
	fmt.Fprint(w, fmt.Sprintf("%s", string(custJSONs)))
}

func Info(w http.ResponseWriter, r *http.Request) {
	var sj map[string]interface{}
	var custJSONs []byte
	obj := r.URL.Query().Get("object")
	id := r.URL.Query().Get("id")

	custJSONs = readJson(Config[obj])
	err = json.Unmarshal(custJSONs, &sj)
	if err != nil {
		log.Println(err)
	}
	j, err := json.Marshal(sj[id])
	if err != nil {
		log.Println(err)
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Server", Config["version"])
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.RequestURI(), "\t", http.StatusOK, "\t", r.UserAgent())
	fmt.Fprint(w, fmt.Sprintf("%s", j))
}

func Update(w http.ResponseWriter, r *http.Request) {
	var (
		sj        map[string]interface{}
		new_cj    interface{}
		err       error
		custJSONs []byte
		o_name    string
		o_id      string
	)
	r.ParseForm()
	queryValues := r.PostFormValue
	id := queryValues("id")
	obj := queryValues("object")
	custJSONs = readJson(Config[obj])
	err = json.Unmarshal(custJSONs, &sj)
	if err != nil {
		log.Println(err)
	}

	if obj == "scheduler" {
		ex := ""
		if queryValues("exclude") == "" {
			ex = "no"
		}
		new_cj = ScheduleJSON{
			Id:        queryValues("id"),
			Name:      queryValues("name"),
			Workday:   r.Form["wd"],
			Stoptime:  queryValues("stoptime"),
			Starttime: queryValues("starttime"),
			Exclude:   ex}
		o_id = queryValues("id")
		o_name = queryValues("name")
	} else if obj == "snapshots" {
		new_cj = SnapshotJSON{
			Id:       queryValues("id"),
			Name:     queryValues("name"),
			Keepdays: helpers.Atoi(queryValues("keepdays"))}
		o_id = queryValues("id")
		o_name = queryValues("name")
	}

	if o_id != "" && o_name != "" {
		sj[id] = new_cj
		jbytes, _ := json.Marshal(sj)
		err = writeJSON(Config[obj], jbytes)
	} else {
		err = fmt.Errorf("Required parameters empty\nData dump:\n%v\n", new_cj)
	}

	if err != nil {
		log.Println(err)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Server", Config["version"])
		log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", r.UserAgent())
		fmt.Fprintf(w, "<h1>Error while file saving</h1><p>Data dump:</p><pre>%v</pre><a href='/admin'>back</a>", new_cj)
	} else {
		log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
		http.Redirect(w, r, "/"+obj, http.StatusMovedPermanently)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	var (
		sj  map[string]interface{}
		err error
		//custJSONs []byte
	)
	r.ParseForm()
	queryValues := r.PostFormValue
	id := queryValues("id")
	obj := queryValues("object")
	//custJSONs = readJson(Config[obj])
	err = json.Unmarshal(readJson(Config[obj]), &sj)
	if err != nil {
		log.Println(err)
	}

	/*_, err := json.Marshal(cj[id])
	if err != nil {
		log.Println(err)
	}
	*/

	delete(sj, id)
	jbytes, _ := json.Marshal(sj)
	err = writeJSON(Config[obj], jbytes)
	if err != nil {
		log.Println(err)
	}

	if err != nil {
		log.Println(err)
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.Header().Set("Server", Config["version"])
		log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusServiceUnavailable, "\t", r.UserAgent())
		fmt.Fprint(w, "<h1>Error while file saving</h1><a href='/admin'>back</a>")
	} else {
		log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
		http.Redirect(w, r, "/"+obj, http.StatusMovedPermanently)
	}
}

func Auth(w http.ResponseWriter, r *http.Request) {
	htp := make(htpass.HTPassFile)
	err := htp.OpenHTPASSFile(Config["pswdfile"])
	queryValues := r.PostFormValue
	res, err := htp.Auth(queryValues("username"), queryValues("passwd"))
	log.Printf("%s\t%v\t%v\t%v\n", r.RemoteAddr, htpass.IsAuth, res, err)
	if err != nil {
		fmt.Println(err)
	}
	authttl, _ = strconv.Atoi(Config["authttl"])
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

func Dump(w http.ResponseWriter, r *http.Request) {
	urlPart := strings.Split(r.URL.Path, "/")
	dumpWhat := urlPart[2]
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", http.StatusOK, "\t", r.UserAgent())
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Server", Config["version"])

	switch dumpWhat {
	case "config":
		{
			fmt.Fprint(w, "<pre><ul>")
			for k, cfg := range Config {
				fmt.Fprintf(w, "<li>%s = %s</li>", k, cfg)
			}
			fmt.Fprint(w, "</ul></pre>")
		}
	case "log":
		{
			respFile, err := os.OpenFile(Config["log_file"], os.O_RDONLY, 0)
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
	base := Config["document_root"]

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
		log.Println(r.RemoteAddr, "\t", err)
		file = "/404.html"
		respFile, err = os.OpenFile(base+file, os.O_RDONLY, 0)
		code = http.StatusNotFound
	}
	/* считать содержимое файла */
	fi, err := respFile.Stat()
	contentType := mime.TypeByExtension(path.Ext(file))
	var bytes = make([]byte, fi.Size())
	respFile.Read(bytes)
	log.Println(r.RemoteAddr, "\t", r.Method, "\t", r.URL.Path, "\t", code, "\t", r.UserAgent())
	/* отправить его клиенту */
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Server", Config["version"])
	w.Write(bytes)
}
