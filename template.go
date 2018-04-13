/*
 * Copyright 2018 darkdarkfruit.  All rights reserved.
 *
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 *
 */

/*
Go(Golang) template manager, especially suited for web.


# There are 2 types of templateEnv(aka: 2 types of templateName). Default is ContextMode
1. ContextMode: Name starts with "C->" or not starts with "F->"

	eg: "C->main/demo/demo.tpl.html"
	 or "C-> main/demo/demo.tpl.html" (not: blanks before file(main/demo/demo.tpl.html) will be discarded)
	 or	"main/demo/demo.tpl.html"

2. FilesMode:   Name starts with "F->". (default separator of multiple files is ";")

	eg: "F->main/demo/demo.tpl.html"
     or "F-> main/demo/demo.tpl.html"
     or "F-> main/demo/demo.tpl.html;main/demo/demo_ads.tpl.html" (will use the first file name when executing template)

# ContextMode is using template nesting, somewhat like template-inheritance in django/jinja2/...
ContextMode will load context templates, then execute template in file: `FilePathOfBaseRelativeToRoot`.

# FilesMode is basically the same as http/template



*/
package templatemanager

//package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"github.com/darkdarkfruit/templatemanager/tplenv"
)

var (
	htmlContentType   = []string{"text/html; charset=utf-8"}
	templateEngineKey = "github.com/foolin/gin-template/templateEngine"
	VERSION = "0.7.0"
)

func Version() string {
	return VERSION
}

type TemplateManager struct {
	Config        TemplateConfig
	TemplatesMap  map[string]*template.Template
	templateMutex sync.RWMutex
}

type TemplateConfig struct {
	DirOfRoot                    string           //template root dir
	DirOfMainRelativeToRoot      string           //template dir: main
	DirOfContextRelativeToRoot   string           //template dir: context
	FilePathOfBaseRelativeToRoot string           //template layout file path
	Extension                    string           //template extension
	FuncMap                      template.FuncMap //template functions
	Delims                       Delims           //delimeters

	IsDebugging bool // true: Show debug info; false: disable debug info and enable cache.
}

type Delims struct {
	Left  string
	Right string
}

func New(config TemplateConfig) *TemplateManager {
	return &TemplateManager{
		Config: config,

		TemplatesMap:  make(map[string]*template.Template),
		templateMutex: sync.RWMutex{},
	}
}
func DefaultConfig(isDebugging bool) TemplateConfig {
	return TemplateConfig{
		DirOfRoot:                    "templates",
		DirOfMainRelativeToRoot:      "main",
		DirOfContextRelativeToRoot:   "context",
		FilePathOfBaseRelativeToRoot: "context/layout/layout.tpl.html",
		Extension:                    ".html",
		FuncMap:                      make(template.FuncMap),
		Delims:                       Delims{Left: "{{", Right: "}}"},
		IsDebugging:                  isDebugging,
	}
}
func Default(isDebugging bool) *TemplateManager {
	return New(DefaultConfig(isDebugging))
}

func (tm *TemplateManager) GetTemplateNames() (names []string) {
	for k := range tm.TemplatesMap {
		names = append(names, k)
	}
	return
}

func (tm *TemplateManager) GetFilePathOfBase() (name string) {
	return path.Join(tm.Config.DirOfRoot, tm.Config.FilePathOfBaseRelativeToRoot)
}

func (tm *TemplateManager) GetMapOfTemplateNameToDefinedNames() (m map[string]string) {
	m = make(map[string]string)
	for k, tpl := range tm.TemplatesMap {
		m[k] = tpl.DefinedTemplates()
	}
	return
}

func (tm *TemplateManager) Report() string {
	s := "Report of template manager: \n"
	s += "============================== \n"
	s += "--> config: \n"
	s += fmt.Sprintf("%#v\n", tm.Config)
	s += "------------------------\n"
	s += fmt.Sprintf("--> (map of templateName -> it's definedNames(%d templateNames total))\n", len(tm.TemplatesMap))
	i := 0
	for tplName, definedNames := range tm.GetMapOfTemplateNameToDefinedNames() {
		i += 1
		s += fmt.Sprintf("%d: %q -> %s\n", i, tplName, definedNames)
	}
	s += "------------------------\n"
	s += "============================== \n"
	return s
}

func (tm *TemplateManager) getDirOfMain() string {
	return path.Join(tm.Config.DirOfRoot, tm.Config.DirOfMainRelativeToRoot)
}

func (tm *TemplateManager) getDirOfContext() string {
	return path.Join(tm.Config.DirOfRoot, tm.Config.DirOfContextRelativeToRoot)
}

func getTemplateFilePathsByWalking(root string, ext string, prefix string) []string {
	var filePaths []string
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() && path.Ext(p) == ext {
			filePaths = append(filePaths, path.Join(prefix, p))
		}
		return nil
	}
	filepath.Walk(root, walkFunc)
	return filePaths
}

// ContainsString checks if the slice has the contains value in it.
func ContainsString(slice []string, contains string) bool {
	for _, value := range slice {
		if value == contains {
			return true
		}
	}
	return false
}

func (tm *TemplateManager) getContextFiles() []string {
	contextFiles := getTemplateFilePathsByWalking(tm.getDirOfContext(), tm.Config.Extension, "")
	if tm.Config.IsDebugging {
		log.Printf("ContextFiles are: %v", contextFiles)
	}
	if !ContainsString(contextFiles, tm.GetFilePathOfBase()) {
		contextFiles = append(contextFiles, tm.GetFilePathOfBase())
	}

	return contextFiles
}

// get templates which is not context file.
func (tm *TemplateManager) getMainFiles() []string {
	//mainFiles, err := filepath.Glob(path.Join(tm.getDirOfMain(), "**", "*"+tm.Config.Extension))
	mainFiles := getTemplateFilePathsByWalking(tm.getDirOfMain(), tm.Config.Extension, "")

	// DirOfContextRelativeToRoot might be a sub directory of DirOfMainRelativeToRoot
	var mf []string
	for _, f := range mainFiles {
		// skip context files (if context_dir is a sub_dir of main_dir)
		if strings.HasPrefix(f, tm.Config.DirOfContextRelativeToRoot) {
			continue
		}
		mf = append(mf, f)
	}

	log.Printf("Found %d main templates(exclude context templates)", len(mf))
	return mf
}

func (tm *TemplateManager) getBasicTemplateNameByFilePath(filepath string) string {
	s := strings.TrimPrefix(filepath, tm.Config.DirOfRoot)
	return strings.TrimPrefix(s, "/")
}

func (tm *TemplateManager) getContextTemplate() *template.Template {
	contextTemplate := template.Must(template.ParseFiles(tm.getContextFiles()...))
	return contextTemplate
}

func (tm *TemplateManager) setTemplate(te *tplenv.TemplateEnv, tpl *template.Template) {
	if tpl == nil {
		panic("Template can not be nil")
	}

	tplName := te.StandardTemplateName()
	tm.templateMutex.Lock()
	defer tm.templateMutex.Unlock()
	tm.TemplatesMap[tplName] = tpl
}

func (tm *TemplateManager) parseMainFiles() error {
	for i, f := range tm.getMainFiles() {
		if tm.Config.IsDebugging {
			log.Printf("--> (seq: %d) Parsing template file: %q", i, f)
		}
		tm.parseMainTemplateByFilePath(f)
	}
	log.Printf("")
	return nil
}

func (tm *TemplateManager) ParseContextModeTemplate(te *tplenv.TemplateEnv) *template.Template {
	if !te.IsContextMode() {
		return nil
	}

	tplName := te.StandardTemplateName()
	filePaths := te.GetFilePaths(tm.Config.DirOfRoot)

	if tm.Config.IsDebugging {
		if len(filePaths) == 1 {
			log.Printf("ContextEnv Parsing: (tplName -> tplPath) (%q -> %q)", tplName, filePaths[0])
		} else {
			log.Printf("ContextEnv Parsing: (tplName -> tplPaths) (%q -> %q)", tplName, filePaths)
		}

	}
	contextFiles := tm.getContextFiles()
	filesForParsing := append(contextFiles, filePaths...)

	tpl := template.Must(template.New(tplName).Funcs(tm.Config.FuncMap).ParseFiles(filesForParsing...))
	tm.setTemplate(te, tpl)
	if tm.Config.IsDebugging {
		log.Printf("ContextEnv template:     (templateName -> definedTemplates): %q -> %s", tpl.Name(), tpl.DefinedTemplates())
	}
	return tpl
}
func (tm *TemplateManager) ParseFilesModeTemplate(te *tplenv.TemplateEnv) *template.Template {
	if !te.IsFilesMode() {
		return nil
	}
	tplName := te.StandardTemplateName()
	filesForParsing := te.GetFilePaths(tm.Config.DirOfRoot)
	if tm.Config.IsDebugging {
		log.Printf("FilesEnv Parsing: (tplName -> tplPath) (%q -> %q)", tplName, filesForParsing)
	}
	tpl := template.Must(template.New(tplName).Funcs(tm.Config.FuncMap).ParseFiles(filesForParsing...))
	tm.setTemplate(te, tpl)
	if tm.Config.IsDebugging {
		log.Printf("FilesEnv template:     (templateName -> definedTemplates): %q -> %s", tpl.Name(), tpl.DefinedTemplates())
	}
	return tpl
}

func (tm *TemplateManager) parseTemplate(te *tplenv.TemplateEnv) *template.Template {
	tplName := te.StandardTemplateName()
	if te.IsContextMode() {
		log.Printf("tplName: %q is a contextEnv tplName", tplName)
		return tm.ParseContextModeTemplate(te)
	} else if te.IsFilesMode() {
		log.Printf("tplName: %q is a filesEnv tplName", tplName)
		return tm.ParseFilesModeTemplate(te)
	} else {
		log.Printf("tplName: %q is an invalid tplName", tplName)
		msg := fmt.Sprintf("Could not find template by tplName: %q", tplName)
		log.Printf(msg)
		panic(msg)
	}
}

func (tm *TemplateManager) parseMainTemplateByFilePath(filePath string) *template.Template {

	basicTplName := tm.getBasicTemplateNameByFilePath(filePath)
	te := tplenv.NewTemplateEnvByParsing(basicTplName)
	te.ToContextMode()
	contextTpl := tm.ParseContextModeTemplate(te)
	if contextTpl == nil {
		log.Printf("Error filepath: %q", filePath)
	}

	te.ToFilesMode()
	filesTpl := tm.ParseFilesModeTemplate(te)
	if filesTpl == nil {
		log.Printf("Error filepath: %q", filePath)
	}
	return contextTpl
}

func (tm *TemplateManager) Init(useMaster bool) error {
	log.Printf("Initing templates. DirOfMainRelativeToRoot: %q, DirOfContextRelativeToRoot: %q", tm.Config.DirOfMainRelativeToRoot, tm.Config.DirOfContextRelativeToRoot)
	includeFunc := func(name string, data interface{}) (template.HTML, error) {
		buf := new(bytes.Buffer)
		err := tm.ExecuteTemplate(buf, name, data)
		return template.HTML(buf.String()), err
	}
	tm.Config.FuncMap["include"] = includeFunc

	return tm.parseMainFiles()

}

func (tm *TemplateManager) GetTemplate(tplName string) (*template.Template, bool) {
	tm.templateMutex.RLock()
	defer tm.templateMutex.RUnlock()
	tpl, ok := tm.TemplatesMap[tplName]
	return tpl, ok
}

func (tm *TemplateManager) ExecuteTemplate(out io.Writer, templateName string, data interface{}) error {
	var tpl *template.Template
	var err error
	var ok bool

	te := tplenv.NewTemplateEnvByParsing(templateName)
	tplName := te.StandardTemplateName()
	if tm.Config.IsDebugging {
		log.Printf("Request executing template name: %q, standard template name is: %q", templateName, tplName)
	}
	tpl, ok = tm.GetTemplate(tplName)

	if !ok || tm.Config.IsDebugging {
		log.Printf("Debug mode. Requst executing templateName: %q. Re-parsing it.", tplName)
		tpl = tm.parseTemplate(te)
		tpl, ok = tm.GetTemplate(tplName)
		if !ok {
			log.Printf("Could not find correspondent template by tplName: %s", tplName)
		}
	}

	// render
	if te.IsContextMode() {
		err = tpl.ExecuteTemplate(out, filepath.Base(tm.Config.FilePathOfBaseRelativeToRoot), data)
		if err != nil {
			log.Printf("TemplateManager execute template error: %s", err)
			return err
		}
	} else if te.IsFilesMode() {
		name := filepath.Base(te.Names[0])
		err = tpl.ExecuteTemplate(out, name, data)
		if err != nil {
			log.Printf("TemplateManager execute template error: %s", err)
			return err
		}
	}

	return nil
}

// ------------------------------
// --->>>-------------------->>>2018-04-10T05-22-39>>>---
// -------- adapt for webserver: gin --------
//
// ---<<<--------------------<<<2018-04-10T05-22-39<<<---

type TemplateRender struct {
	templateManager *TemplateManager
	Name            string
	Data            interface{}
}

func (tm *TemplateManager) Instance(name string, data interface{}) render.Render {
	return TemplateRender{
		templateManager: tm,
		Name:            name,
		Data:            data,
	}
}

func (tm *TemplateManager) executeRender(out io.Writer, name string, data interface{}) error {
	return tm.ExecuteTemplate(out, name, data)
}

func (r TemplateRender) Render(w http.ResponseWriter) error {
	return r.templateManager.executeRender(w, r.Name, r.Data)
}

func (r TemplateRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = htmlContentType
	}
}

func (tm *TemplateManager) HTML(ctx *gin.Context, code int, name string, data interface{}) {
	instance := tm.Instance(name, data)
	ctx.Render(code, instance)
}

func HTML(ctx *gin.Context, code int, name string, data interface{}) {
	if val, ok := ctx.Get(templateEngineKey); ok {
		if tm, ok := val.(*TemplateManager); ok {
			tm.HTML(ctx, code, name, data)
			return
		}
	}
	ctx.HTML(code, name, data)
}

//func NewMiddleware(config TemplateConfig) gin.HandlerFunc {
//	return Middleware(New(config))
//}
//
//func Middleware(tm *TemplateManager) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		c.Set(templateEngineKey, tm)
//	}
//}
