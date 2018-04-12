package main

import (
	"log"
	"time"
	"os"
	"github.com/darkdarkfruit/templatemanager/tplenv"
	"github.com/darkdarkfruit/templatemanager"
)

func main() {
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	//tplConf := DefaultConfig(true)
	//tplMgr := New(tplConf)
	tplMgr := templatemanager.Default(true)
	tplMgr.Init(true)
	log.Printf(tplMgr.Report())
	log.Printf("templateNames are: %s", tplMgr.GetTemplateNames())
	tplName := "main/demo/demo.tpl.html"
	data := map[string]interface{}{
		"tplName": tplName,
		"tplPath": "",
		"now":     time.Now(),
	}
	log.Printf(tplMgr.Report())
	time.Sleep(time.Millisecond * 500)

	log.Println("ContextEnv: render a file")
	err = tplMgr.ExecuteTemplate(os.Stdout, tplName, data)
	if err != nil {
		log.Printf("%s", err)
	}

	//log.Println("ContextEnv: render two files")
	//tplName += ";main/demo/demo2.tpl.html"
	//time.Sleep(time.Millisecond * 200)
	//err = tplMgr.executeTemplate(os.Stdout, tplName, data)
	//if err != nil {
	//	log.Printf("%s", err)
	//}

	log.Printf("FilesEnv: render a file")
	singleTplName := tplenv.NewTemplateEnvByParsing(tplName).ToFilesMode().StandardTemplateName()
	err = tplMgr.ExecuteTemplate(os.Stdout, singleTplName, data)
	if err != nil {
		log.Printf("%s", err)
	}

	log.Printf("FilesEnv: render 2 files")
	tplName = string(tplenv.TemplateModeFilesPrefix) + " main/demo/demo2.tpl.html;main/demo/demo.tpl.html"
	err = tplMgr.ExecuteTemplate(os.Stdout, tplName, data)
	if err != nil {
		log.Printf("%s", err)
	}

	time.Sleep(time.Millisecond * 300)
	err = tplMgr.ExecuteTemplate(os.Stdout, "main/demo/dir1/dir2/any.tpl.html", data)
	if err != nil {
		log.Printf("%s", err)
	}


	err = tplMgr.ExecuteTemplate(os.Stdout, "F->main/demo/dir1/dir2/any.tpl.html", data)
	if err != nil {
		log.Printf("%s", err)
	}

	//tplName = "template/main/demo/demo2.tpl.html"
	//data = map[string]interface{}{
	//	"tplName": tplName,
	//	"tplPath": "",
	//	"now":     time.Now(),
	//}
	//time.Sleep(time.Millisecond * 500)
	//err = tplMgr.executeTemplate(os.Stdout, tplName, data)
	//if err != nil {
	//	log.Printf("%s", err)
	//}
	////time.Sleep(time.Second)
	//
}

