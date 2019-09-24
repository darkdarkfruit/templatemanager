package templatemanager

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"io"
	"net/http"
)

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

//func (r TemplateRender) _Render(w http.ResponseWriter) error {
//	return r.templateManager.executeRender(w, r.Name, r.Data)
//}

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
