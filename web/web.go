package web

import (
	"log"
	"net/http"

	"shortme/conf"

	"os"
	"shortme/web/api"
	"shortme/web/www"

	"github.com/gorilla/mux"
)

var logger *log.Logger

func Start() {
	log.Println("web starts")
	r := mux.NewRouter()

	r.HandleFunc("/version", api.CheckVersion).Methods(http.MethodGet)
	r.HandleFunc("/health", api.CheckHealth).Methods(http.MethodGet)
	r.HandleFunc("/short", api.ShortURL).Methods(http.MethodPost).HeadersRegexp("Content-Type", "application/json")
	r.HandleFunc("/expand", api.ExpandURL).Methods(http.MethodPost).HeadersRegexp("Content-Type", "application/json")
	r.HandleFunc("/{shortenedURL:[a-zA-Z0-9]{1,11}}", api.Redirect).Methods(http.MethodGet)

	r.HandleFunc("/index.html", www.Index).Methods(http.MethodGet)

	r.Handle("/static/{type}/{file}", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	r.Handle("/favicon.ico", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	file, err := os.OpenFile("test.log", os.O_APPEND|os.O_CREATE, 666)
	if err != nil {
		log.Fatalln("fail to create test.log file!")
	}
	defer file.Close()
	logger = log.New(file, "", log.LstdFlags|log.Lshortfile) // 日志文件格式:log包含时间及文件行数
	log.Println("输出日志到命令行终端")
	logger.Println("将日志写入文件")

	logger.SetFlags(log.LstdFlags | log.Lshortfile) // 设置日志格式

	log.Fatal(http.ListenAndServe(conf.Conf.Http.Listen, r))
}
