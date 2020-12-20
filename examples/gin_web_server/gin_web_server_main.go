package main

import (
	"github.com/darkdarkfruit/templatemanager"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"time"
)

func time_isoformat(t time.Time) string {
	return t.Format(time.RFC3339)
}

func HomeHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "main/home/home.tpl.html", render.M{
		"now": time.Now(),
	})
}

func Demo1Handler(c *gin.Context) {
	c.HTML(http.StatusOK, "main/demo/demo1.tpl.html", render.M{
		"now": time.Now(),
	})
}

func Demo2Handler(c *gin.Context) {
	c.HTML(http.StatusOK, "main/demo/demo2.tpl.html", render.M{
		"now": time.Now(),
	})
}

func AnyFileHandler(c *gin.Context) {
	tplName := "main/demo/dir1/dir2/any.tpl.html"
	c.HTML(http.StatusOK, tplName, render.M{
		"now":     time.Now(),
		"tplName": tplName,
	})
}

func main() {
	log.Printf("gin mode: %s, isDebugging: %v", gin.Mode(), gin.IsDebugging())

	router := gin.Default()
	tplMgr := templatemanager.NewDefault(false)
	tplMgr.Config.FuncMap = template.FuncMap{
		"time_isoformat": time_isoformat,
	}
	tplMgr.Init(true)
	router.HTMLRender = tplMgr

	router.GET("/", HomeHandler)
	router.GET("/demo1", Demo1Handler)
	router.GET("/demo2", Demo2Handler)
	router.GET("/any", AnyFileHandler)

	addr := ":10000"
	log.Printf("gin-web-server is running at %s", addr)
	router.Run(addr)
}
