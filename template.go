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
ContextMode will load context templates, then execute template in file: `FilePathOfLayoutRelativeToRoot`.

# FilesMode is basically the same as http/template

*/
package templatemanager

// package main

import (
	"bytes"
	"fmt"
	"github.com/oxtoacart/bpool"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	MimeHtml          = "text/html"
	templateEngineKey = "github.com/darkdarkfruit/templatemanager"
	VERSION           = "0.7.1"
)

var (
	htmlContentType    = []string{"text/html; charset=utf-8"}
	bufpool            *bpool.BufferPool
	htmlMinifier       *minify.M
	goTemplateMinifier *minify.M
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	bufpool = bpool.NewBufferPool(64)
	// log.Println("buffer allocation successful")
	htmlMinifier = minify.New()
	htmlMinifier.AddFunc(MimeHtml, html.Minify)

	goTemplateMinifier = minify.New()

	// we keep the minifier wild.
	goTemplateMinifier.Add(MimeHtml, &html.Minifier{
		KeepDefaultAttrVals: true,
		KeepWhitespace:      true,
		KeepDocumentTags:    true,
		KeepEndTags:         true,
	})
}

func Version() string {
	return VERSION
}

type TemplateManager struct {
	Config        TemplateConfig
	TemplatesMap  map[string]*template.Template
	templateMutex sync.RWMutex
}

type TemplateConfig struct {
	DirOfRoot                      string           // template root dir
	DirOfMainRelativeToRoot        string           // template dir: main
	DirOfContextRelativeToRoot     string           // template dir: context
	FilePathOfLayoutRelativeToRoot string           // template layout file path
	Extension                      string           // template extension
	FuncMap                        template.FuncMap // template functions
	Delims                         Delims           // delimiters

	IsDebugging          bool // true: Show debug info; false: disable debug info and enable cache.
	VerboseLevel         int  // 0: not show anything
	EnableMinifyTemplate bool // enable minify template after loading it and before storing it to the memory.
	EnableMinifyHtml     bool // decide to minify html while output
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

func NewDefault(isDebugging bool) *TemplateManager {
	return New(NewDefaultConfig(isDebugging))
}

func NewDefaultConfig(isDebugging bool) TemplateConfig {
	return TemplateConfig{
		DirOfRoot:                      "templates",
		DirOfMainRelativeToRoot:        "main",
		DirOfContextRelativeToRoot:     "context",
		FilePathOfLayoutRelativeToRoot: "context/layout/layout.tpl.html",
		Extension:                      ".html",
		FuncMap:                        make(template.FuncMap),
		Delims:                         Delims{Left: "{{", Right: "}}"},
		IsDebugging:                    isDebugging,
		VerboseLevel:                   1,
		EnableMinifyTemplate:           false,
		EnableMinifyHtml:               false,
	}
}
func (tm *TemplateManager) DoShowDebugMessage() bool {
	if tm.Config.VerboseLevel <= 0 {
		return false
	}
	return tm.Config.IsDebugging && tm.Config.VerboseLevel > 0
}

// 0: disables all
func (tm *TemplateManager) SetVerboseLevel(level int) {
	tm.Config.VerboseLevel = level
}

func (tm *TemplateManager) GetTemplateNames() (names []string) {
	for k := range tm.TemplatesMap {
		names = append(names, k)
	}
	return
}

func (tm *TemplateManager) GetFilePathOfBase() (name string) {
	return path.Join(tm.Config.DirOfRoot, tm.Config.FilePathOfLayoutRelativeToRoot)
}

func (tm *TemplateManager) GetMapOfTemplateNameToDefinedNames() (m map[string]string) {
	m = make(map[string]string)
	for k, tpl := range tm.TemplatesMap {
		m[k] = tpl.DefinedTemplates()
	}
	return
}

func (tm *TemplateManager) Report() string {
	s := fmt.Sprintf(`
Report of template manager
==============================
--> config
%#v
------------------------
--> (map(sum=%d):  templateName -> it's definedNames), 
`, tm.Config, len(tm.TemplatesMap))
	i := 0
	for tplName, definedNames := range tm.GetMapOfTemplateNameToDefinedNames() {
		i += 1
		s += fmt.Sprintf("%d: %q -> %s\n", i, tplName, definedNames)
	}
	s += "------------------------\n"
	return s
}

func (tm *TemplateManager) getDirOfMain() string {
	return path.Join(tm.Config.DirOfRoot, tm.Config.DirOfMainRelativeToRoot)
}

func (tm *TemplateManager) getDirOfContext() string {
	return path.Join(tm.Config.DirOfRoot, tm.Config.DirOfContextRelativeToRoot)
}

func getTemplateFilePathsByWalking(root string, ext string, prefix string) ([]string, error) {
	var filePaths []string
	walkFunc := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			currentDir, _ := os.Getwd()
			log.Panicf("error happens while walking dir: %q(current dir is: %q), err: %v", p, currentDir, err)
		}
		if !info.IsDir() && path.Ext(p) == ext {
			filePaths = append(filePaths, path.Join(prefix, p))
		}
		return nil
	}
	err := filepath.Walk(root, walkFunc)
	if err != nil {
		log.Printf("Faild walking root dir: %q. err: %q", root, err)
		log.Fatalf("Faild walking root dir: %q. err: %q", root, err)
		return nil, err
	}
	return filePaths, nil
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
	contextFiles, err := getTemplateFilePathsByWalking(tm.getDirOfContext(), tm.Config.Extension, "")
	if err != nil {
		log.Fatalf("Could not get context files of dir: %q. err: %s", tm.getDirOfContext(), err)
	}
	if tm.DoShowDebugMessage() {
		log.Printf("ContextFiles are: %v", contextFiles)
	}
	if !ContainsString(contextFiles, tm.GetFilePathOfBase()) {
		contextFiles = append(contextFiles, tm.GetFilePathOfBase())
	}

	return contextFiles
}

// get templates which is not context file.
func (tm *TemplateManager) getMainFiles() []string {
	// mainFiles, err := filepath.Glob(path.Join(tm.getDirOfMain(), "**", "*"+tm.Config.Extension))
	mainFiles, err := getTemplateFilePathsByWalking(tm.getDirOfMain(), tm.Config.Extension, "")
	if err != nil {
		log.Fatalf("Could not get main files of dir: %q. err: %s", tm.getDirOfMain(), err)
	}

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

// func (tm *TemplateManager) getContextTemplate() *template.Template {
//	contextTemplate := template.Must(template.ParseFiles(tm.getContextFiles()...))
//	return contextTemplate
//}

func (tm *TemplateManager) setTemplate(te *TemplateEnv, tpl *template.Template) {
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
		if tm.DoShowDebugMessage() {
			log.Printf("\n")
			log.Printf("--(template: seq: %d)--> Parsing template file: %q", i, f)
		}
		tm.parseMainTemplateByFilePath(f)
	}
	log.Printf("")
	return nil
}

func (tm *TemplateManager) MustTemplate(tplName string, filesForParsing []string) *template.Template {
	if !tm.Config.EnableMinifyTemplate {
		tpl := template.Must(template.New(tplName).Funcs(tm.Config.FuncMap).ParseFiles(filesForParsing...))
		return tpl
	} else {
		/* Could not find a better way to store minified content while keeps filenames which required by template to "ParseFiles"
		 * solution:
		 * 	create a tmp dir, then populates it with minified files.
		 * ...
		 * ...
		 */
		if tm.DoShowDebugMessage() {
			log.Printf("Minifying template: %q", tplName)
		}
		tmpDir, err := ioutil.TempDir("", "go-template")
		if err != nil {
			log.Printf("Could not create temparary dir. err: %s", err)
			panic(err)
			return nil
		}
		defer func() {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				log.Printf("could not remove tmpDir: %q. e: %v", tmpDir, err)
			}
		}()

		for _, f := range filesForParsing {
			baseName := path.Base(f)
			fpath := path.Join(tmpDir, baseName)
			outFile, err := os.Create(fpath)
			if err != nil {
				fmt.Printf("Could not create file: %q. err: %s", fpath, err)
				panic(err)
				return nil
			}

			inFile, err := os.Open(f)
			if err != nil {
				fmt.Printf("Could not open file: %q. err: %s", f, err)
				panic(err)
				return nil
			}
			err = goTemplateMinifier.Minify(MimeHtml, outFile, inFile)
			if err != nil {
				log.Printf("Could not minify template in buf. err: %s", err)
				panic(err)
			}
		}
		tpl := template.Must(template.New(tplName).Funcs(tm.Config.FuncMap).ParseGlob(filepath.Join(tmpDir, "*")))
		return tpl
	}
}

func (tm *TemplateManager) ParseContextModeTemplate(te *TemplateEnv) *template.Template {
	if !te.IsContextMode() {
		return nil
	}

	tplName := te.StandardTemplateName()
	filePaths := te.GetFilePaths(tm.Config.DirOfRoot)

	if tm.DoShowDebugMessage() {
		if len(filePaths) == 1 {
			log.Printf("ContextEnv Parsing: (tplName -> tplPath) (%q -> %q)", tplName, filePaths[0])
		} else {
			log.Printf("ContextEnv Parsing: (tplName -> tplPaths) (%q -> %q)", tplName, filePaths)
		}
	}
	contextFiles := tm.getContextFiles()
	filesForParsing := append(contextFiles, filePaths...)

	// tpl := template.Must(template.New(tplName).Funcs(tm.Config.FuncMap).ParseFiles(filesForParsing...))
	tpl := tm.MustTemplate(tplName, filesForParsing)
	tm.setTemplate(te, tpl)
	if tm.DoShowDebugMessage() {
		log.Printf("ContextEnv template:     (templateName -> definedTemplates): %q -> %s", tpl.Name(), tpl.DefinedTemplates())
	}
	return tpl
}

func (tm *TemplateManager) ParseFilesModeTemplate(te *TemplateEnv) *template.Template {
	if !te.IsFilesMode() {
		return nil
	}
	tplName := te.StandardTemplateName()
	filesForParsing := te.GetFilePaths(tm.Config.DirOfRoot)
	if tm.DoShowDebugMessage() {
		log.Printf("FilesEnv Parsing: (tplName -> tplPath) (%q -> %q)", tplName, filesForParsing)
	}
	// tpl := template.Must(template.New(tplName).Funcs(tm.Config.FuncMap).ParseFiles(filesForParsing...))
	tpl := tm.MustTemplate(tplName, filesForParsing)
	tm.setTemplate(te, tpl)
	if tm.DoShowDebugMessage() {
		log.Printf("FilesEnv template: (tplName -> definedTemplates): %q -> %s", tpl.Name(), tpl.DefinedTemplates())
	}
	return tpl
}

func (tm *TemplateManager) parseTemplate(te *TemplateEnv) *template.Template {
	tplName := te.StandardTemplateName()
	if te.IsContextMode() {
		if tm.DoShowDebugMessage() {
			log.Printf("tplName: %q is a contextEnv tplName", tplName)
		}
		return tm.ParseContextModeTemplate(te)
	} else if te.IsFilesMode() {
		if tm.DoShowDebugMessage() {
			log.Printf("tplName: %q is a filesEnv tplName", tplName)
		}
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
	te := NewTemplateEnvByParsing(basicTplName)
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
	t0 := time.Now()
	var tpl *template.Template
	var err error
	var ok bool

	te := NewTemplateEnvByParsing(templateName)
	tplName := te.StandardTemplateName()
	if tm.DoShowDebugMessage() {
		log.Printf("Request executing template name: %q, standard template name is: %q", templateName, tplName)
	}
	tpl, ok = tm.GetTemplate(tplName)

	if !ok || tm.Config.IsDebugging {
		log.Printf("Template-not-exist or in-debug-mode. Requst executing templateName: %q. Re-parsing it.", tplName)
		tpl = tm.parseTemplate(te)
		tpl, ok = tm.GetTemplate(tplName)
		if !ok {
			log.Printf("Could not find correspondent template by tplName: %s", tplName)
		}
	}

	// render
	if te.IsContextMode() {
		err = tpl.ExecuteTemplate(out, filepath.Base(tm.Config.FilePathOfLayoutRelativeToRoot), data)
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

	if tm.Config.VerboseLevel >= 1  {
		log.Printf("ExecuteTemplate stat: %s", NewQpsStat(t0, time.Now(), 1).ShortString())
	}
	return nil
}
