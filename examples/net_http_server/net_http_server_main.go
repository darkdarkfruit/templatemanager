package main

import (
	"github.com/darkdarkfruit/templatemanager"
	"html/template"
	"log"
	"net/http"
	"time"
)

var isDebugging bool
var tplMgr *templatemanager.TemplateManager

//func init() {
//}

func time_isoformat(t time.Time) string {
	return t.Format(time.RFC3339)
}

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	tplMgr.ExecuteTemplate(w, "main/home/home.tpl.html", map[string]interface{}{
		"now": time.Now(),
	})
}

func Demo1Handler(w http.ResponseWriter, req *http.Request) {
	tplMgr.ExecuteTemplate(w, "main/demo/demo1.tpl.html", map[string]interface{}{
		"now": time.Now(),
	})
}

func Demo2Handler(w http.ResponseWriter, req *http.Request) {
	tplMgr.ExecuteTemplate(w, "main/demo/demo2.tpl.html", map[string]interface{}{
		"now": time.Now(),
	})
}

func AnyFileHandler(w http.ResponseWriter, req *http.Request) {
	tplName := "main/demo/dir1/dir2/any.tpl.html"
	tplMgr.ExecuteTemplate(w, tplName, map[string]interface{}{
		"now":     time.Now(),
		"tplName": tplName,
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	isDebugging = true
	tplMgr = templatemanager.NewDefault(isDebugging)
	tplMgr.Config.FuncMap = template.FuncMap{
		"time_isoformat": time_isoformat,
	}
	tplMgr.Init(true)

	mux := http.NewServeMux()
	log.Printf("isDebugging: %v", isDebugging)

	mux.HandleFunc("/", HomeHandler)
	mux.HandleFunc("/demo1", Demo1Handler)
	mux.HandleFunc("/demo2", Demo2Handler)
	mux.HandleFunc("/any", AnyFileHandler)

	addr := ":10001"
	httpAddr := "http://localhost" + addr
	log.Printf("urls are: \n%s/ \n%s/demo1 \n%s/demo2 \n%s/any \n", httpAddr, httpAddr, httpAddr, httpAddr)
	log.Printf("net-http-server is running at %s", addr)
	http.ListenAndServe(addr, mux)
}
