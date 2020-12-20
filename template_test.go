package templatemanager

import (
	"bytes"
	"log"
	"testing"
	"time"
)

var gTplMgr *TemplateManager

func init() {
	//log.SetFlags(log.LstdFlags | log.Lshortfile)
	//tplConf := DefaultConfig(true)
	//gTplMgr := New(tplConf)
	gTplMgr = NewDefault(true)
	err := gTplMgr.Init(true)
	if err != nil {
		panic(err)
	}
	//gTplMgr.SetVerboseLevel(0)
	//gTplMgr.SetVerboseLevel()
	log.Printf("silent: %v", gTplMgr.Config.VerboseLevel)
	//return
	log.Printf(gTplMgr.Report())
	log.Printf("templateNames are: %s\n\n\n", gTplMgr.GetTemplateNames())
}

func TestTemplateManager_ExecuteTemplate(t *testing.T) {
	tplName := "main/demo/demo1.tpl.html"
	data := map[string]interface{}{
		"tplName": tplName,
		"tplPath": "",
		"now":     time.Now(),
	}

	type args struct {
		templateName string
		data         interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		{
			name: "ContextEnv: render a file",
			args: args{
				templateName: "main/demo/demo1.tpl.html",
				data:         data,
			},
			wantErr: false,
		},
		{
			name: "ContextEnv: render any template at any directory depth",
			args: args{
				templateName: "main/demo/dir1/dir2/any.tpl.html",
				data:         data,
			},
			wantErr: false,
		},
		{
			name: "FilesMode: render any template at any directory depth",
			args: args{
				templateName: "F->main/demo/dir1/dir2/any.tpl.html",
				data:         data,
			},
			wantErr: false,
		},
		{
			name: "FilesMode: render multiple template at any directory depth",
			args: args{
				templateName: "F->main/demo/demo2.tpl.html;main/demo/demo1.tpl.html",
				data:         data,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := gTplMgr
			out := &bytes.Buffer{}
			err := tm.ExecuteTemplate(out, tt.args.templateName, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotStr := out.String()
			if len(gotStr) < 10 {
				t.Errorf("ExecuteTemplate() gotOut str length < 100, got: %d", len(gotStr))
				t.Errorf("got: \n%s\n", gotStr)
				return
			}
		})
	}
}

func BenchmarkTemplateManager_ExecuteTemplate(b *testing.B) {
	gTplMgr.Config.IsDebugging = false
	gTplMgr.SetVerboseLevel(0)
	tplName1 := "main/demo/demo1.tpl.html"
	tplName2 := "main/demo/dir1/dir2/any.tpl.html"
	data := map[string]interface{}{
		"tplName1": tplName1,
		"tplPath":  "",
		"now":      time.Now(),
	}

	b.Run("default context mode: main/demo/demo1.tpl.html", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out := &bytes.Buffer{}
			err := gTplMgr.ExecuteTemplate(out, tplName1, data)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("default context mode(deep level file): main/demo/dir1/dir2/any.tpl.html", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out := &bytes.Buffer{}
			err := gTplMgr.ExecuteTemplate(out, tplName2, data)
			if err != nil {
				panic(err)
			}
		}
	})

	b.Run("file mode: main/demo/dir1/dir2/any.tpl.html", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out := &bytes.Buffer{}
			err := gTplMgr.ExecuteTemplate(out, "F->"+tplName2, data)
			if err != nil {
				panic(err)
			}
		}
	})

}
