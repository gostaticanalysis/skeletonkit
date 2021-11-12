package skeletonkit

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"text/template"

	"github.com/josharian/txtarfs"
	"golang.org/x/tools/imports"
	"golang.org/x/tools/txtar"
)

// DefaultFuncMap is a default functions which using in a template.
var DefaultFuncMap = template.FuncMap{
	"gitkeep": func() string {
		return ".gitkeep"
	},
	"gomod": func() string {
		return "go.mod"
	},
	"gomodinit": func(path string) string {
		f, err := ModInit(path)
		if err != nil {
			panic(err)
		}
		return f
	},
}

// TemplateOption is an option for a template.
// It can decorate the template.
type TemplateOption func(*template.Template) (*template.Template, error)

// TemplateWithFuncs add funcs to a template.
func TemplateWithFuncs(funcs template.FuncMap) TemplateOption {
	return func(tmpl *template.Template) (*template.Template, error) {
		return tmpl.Funcs(funcs), nil
	}
}

// TemplateWithDelims sets delims of a template.
func TemplateWithDelims(left, right string) TemplateOption {
	return func(tmpl *template.Template) (*template.Template, error) {
		return tmpl.Delims(left, right), nil
	}
}

// ParseTemplate parses a template which represents as txtar format from given a file system.
// The name is template name. If stripPrefix is not empty string, ParseTemplate uses sub directory of stripPrefix.
func ParseTemplate(tmplFS fs.FS, name, stripPrefix string, options ...TemplateOption) (*template.Template, error) {
	fsys := tmplFS
	if stripPrefix != "" {
		var err error
		fsys, err = fs.Sub(tmplFS, stripPrefix)
		if err != nil {
			return nil, err
		}
	}

	ar, err := txtarfs.From(fsys)
	if err != nil {
		return nil, err
	}

	tmpl := template.New(name).Delims("@@", "@@").Funcs(DefaultFuncMap)
	for _, opt := range options {
		tmpl, err = opt(tmpl)
		if err != nil {
			return nil, err
		}
	}

	return tmpl.Parse(string(txtar.Format(ar)))
}

// ExecuteTemplate executes a template with data.
// ExecuteTemplate also parses the excuted string as a txtar format and return it as a file system.
func ExecuteTemplate(tmpl *template.Template, data interface{}) (fs.FS, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	ar := txtar.Parse(buf.Bytes())
	for i := range ar.Files {
		if filepath.Ext(ar.Files[i].Name) != ".go" ||
			len(bytes.TrimSpace(ar.Files[i].Data)) == 0 {
			continue
		}
		opt := &imports.Options{
			Comments:   true,
			FormatOnly: true,
		}
		src, err := imports.Process(ar.Files[i].Name, ar.Files[i].Data, opt)
		if err != nil {
			return nil, err
		}
		ar.Files[i].Data = src
	}

	return txtarfs.As(ar), nil
}
