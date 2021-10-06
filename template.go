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

// ParseTemplate parses a template which represents as txtar format from given a file system.
// The name is template name. If stripPrefix is not empty string, ParseTemplate uses sub directory of stripPrefix.
func ParseTemplate(tmplFS fs.FS, name, stripPrefix string) (*template.Template, error) {
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

	return template.New(name).Delims("@@", "@@").Funcs(template.FuncMap{
		"gitkeep": func() string {
			return ".gitkeep"
		},
		"gomod": func() string {
			return "go.mod"
		},
		"gomodinit": func(path string) string {
			f, err := ModeInit(path)
			if err != nil {
				panic(err)
			}
			return f
		},
	}).Parse(string(txtar.Format(ar)))
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
