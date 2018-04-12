package main

import (
	"log"
	"time"
	"github.com/darkdarkfruit/templatemanager/tplenv"
	"github.com/darkdarkfruit/templatemanager"
	"bufio"
	"bytes"
)

var cnt  = 0



func executeTemplate(tplMgr *templatemanager.TemplateManager, tplName string, data map[string]interface{}) (*templatemanager.TemplateManager){
	cnt += 1
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err := tplMgr.ExecuteTemplate(writer, tplName, data)
	if err != nil {
		log.Printf("%s", err)
	}
	writer.Flush()
	log.Printf("\n====%d:start====\n%s\n====%d: end ====\n\n\n", cnt, b.String(), cnt)
	return tplMgr
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//tplConf := DefaultConfig(true)
	//tplMgr := New(tplConf)
	tplMgr := templatemanager.Default(true)
	tplMgr.Init(true)
	log.Printf(tplMgr.Report())
	log.Printf("templateNames are: %s\n\n\n", tplMgr.GetTemplateNames())
	tplName := "main/demo/demo1.tpl.html"
	data := map[string]interface{}{
		"tplName": tplName,
		"tplPath": "",
		"now":     time.Now(),
	}
	//log.Printf(tplMgr.Report())
	//delayOutput()
	log.Printf("ContextEnv: render a file: %s", tplName)
	executeTemplate(tplMgr, tplName, data)

	tplName = "main/demo/dir1/dir2/any.tpl.html"
	log.Printf("ContextEnv: render any template at any directory depth: %s", tplName)
	executeTemplate(tplMgr, tplName, data)

	singleTplName := tplenv.NewTemplateEnvByParsing(tplName).ToFilesMode().StandardTemplateName()
	log.Printf("FilesEnv: render a file: %s", tplName)
	executeTemplate(tplMgr, singleTplName, data)

	tplName = string(tplenv.TemplateModeFilesPrefix) + " main/demo/demo2.tpl.html;main/demo/demo1.tpl.html"
	log.Printf("FilesEnv: render 2 files: %s", tplName)
	executeTemplate(tplMgr, tplName, data)

}
