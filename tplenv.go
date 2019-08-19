package templatemanager

import (
	"path"
	"strings"
)

type TemplateModePrefix string

const TemplateModeContextPrefix TemplateModePrefix = "C->"
const TemplateModeFilesPrefix TemplateModePrefix = "F->"
const FilesSeparator = ";"

type TemplateEnv struct {
	Mode  TemplateModePrefix // template env: "C->" or "F->"
	Names []string           // template names. ContextEnv has one "Names" only.
}

func (self TemplateEnv) String() string {
	return self.StandardTemplateName()
}

func (self *TemplateEnv) StandardTemplateName() string {
	s := string(self.Mode)
	return s + strings.Join(self.Names, FilesSeparator)
}

func (self *TemplateEnv) ToContextMode() *TemplateEnv {
	self.Mode = TemplateModeContextPrefix
	return self
}

func (self *TemplateEnv) ToFilesMode() *TemplateEnv {
	self.Mode = TemplateModeFilesPrefix
	return self
}

func getFilesFromTemplateName(tplName, prefix, separator string) []string {
	_tplName := strings.Trim(tplName, " ")
	names := strings.TrimPrefix(_tplName, prefix)
	namesSlice := strings.Split(names, separator)
	var standarizedNamesSlice []string
	for _, v := range namesSlice {
		standarizedNamesSlice = append(standarizedNamesSlice, strings.Trim(v, " "))
	}
	return standarizedNamesSlice
}
func NewTemplateEnvByParsing(tplName string) *TemplateEnv {
	ctxPrefix := string(TemplateModeContextPrefix)
	filesPrefix := string(TemplateModeFilesPrefix)
	if strings.HasPrefix(tplName, ctxPrefix) {
		return &TemplateEnv{
			Mode:  TemplateModeContextPrefix,
			Names: getFilesFromTemplateName(tplName, ctxPrefix, FilesSeparator),
		}

	} else if strings.HasPrefix(tplName, filesPrefix) {
		return &TemplateEnv{
			Mode:  TemplateModeFilesPrefix,
			Names: getFilesFromTemplateName(tplName, filesPrefix, FilesSeparator),
		}
	} else {
		return &TemplateEnv{
			Mode:  TemplateModeContextPrefix,
			Names: getFilesFromTemplateName(tplName, ctxPrefix, FilesSeparator),
		}
	}
}

func (self *TemplateEnv) IsFilesMode() bool {
	return self.Mode == TemplateModeFilesPrefix
}

func (self *TemplateEnv) IsContextMode() bool {
	return self.Mode == TemplateModeContextPrefix || !self.IsFilesMode()
}

func (self *TemplateEnv) GetFilePaths(dir string) []string {
	var paths []string
	for _, name := range self.Names {
		_path := path.Join(dir, name)
		paths = append(paths, _path)
	}
	return paths
}
