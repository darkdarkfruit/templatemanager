Go(Golang) template manager, especially suited for web. Already supports [gin](https://github.com/gin-gonic/gin) server.
========================================================
1. [Install](#install)
2. [Description](#description)
3. [Examples](#examples)


## Install 
go get github.com/darkdarkfruit/templatemanager


## Description


### There are 2 types of templateEnv(aka: 2 types of templateName). 
Default is ContextMode which uses template nesting(somewhat like template-inheritance in django/jinja2/...)

ContextMode will load context templates, then execute template in file: `FilePathOfBaseRelativeToRoot`.

FilesMode is basically the same as http/template

1. ContextMode: Name starts with "C->" or not starts with "F->"
```
	eg: "C->main/demo/demo.tpl.html"
	 or "C-> main/demo/demo.tpl.html" (not: blanks before file(main/demo/demo.tpl.html) will be discarded)
	 or	"main/demo/demo.tpl.html"
```
2. FilesMode:   Name starts with "F->". (default separator of multiple files is ";")
```
	eg: "F->main/demo/demo.tpl.html"
     or "F-> main/demo/demo.tpl.html"
     or "F-> main/demo/demo.tpl.html;main/demo/demo_ads.tpl.html" (will use the first file name when executing template)
```

## Examples
See detailed examples at [examples/](examples)

### We will use the templates layout below:
```

templates/
├── context
│   ├── layout
│   │   └── layout.tpl.html
│   └── partial
│       └── ads.tpl.html
└── main
    └── demo
        ├── demo1.tpl.html
        ├── demo2.tpl.html
        └── dir1
            └── dir2
                └── any.tpl.html

```

#### A very basic example.
```
// use gin as web server
...

func main(){
    router := gin.Default() // 
	tplMgr := templatemanager.Default(true)
	tplMgr.Init(true)
	router.HTMLRender = tplMgr 
	}

...
	
```

#### A customized example
```
// use gin as web server
...

func main(){
    router := gin.Default() // 
    tplConfig := templatemanager.DefaultConfig(gin.IsDebugging())
	tplConfig.DirOfRoot = templatesDir
	tplConfig.FuncMap = template.FuncMap{
			"FormatAsDate":   FormatAsDate,
			"time_ISOFormat": TimeISOFormat,
			"unescaped":      unescaped,
		}
	tplMgr := templatemanager.New(tplConfig)
	tplMgr.Init(true)
	router.HTMLRender = tplMgr 
	}

...
	
```


