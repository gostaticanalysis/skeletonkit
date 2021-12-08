package skeletonkit_test

import (
	"embed"
	"flag"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/gostaticanalysis/skeletonkit"
	"github.com/tenntenn/golden"
)

//go:embed testdata/template
var testTmplFS embed.FS

var (
	flagUpdate bool
)

func init() {
	flag.BoolVar(&flagUpdate, "update", false, "update golden files")
}

func TestFsysCreateDir(t *testing.T) {
	t.Parallel()
	F := golden.TxtarWith
	T := func(opts ...skeletonkit.TemplateOption) []skeletonkit.TemplateOption {
		return opts
	}
	C := func(opts ...skeletonkit.CreatorOption) []skeletonkit.CreatorOption {
		return opts
	}

	type appInfo struct {
		Name       string
		ModulePath string
	}

	type errs struct {
		parse  bool
		exec   bool
		create bool
	}

	cases := map[string]struct {
		dirinit string
		root    string
		info    interface{}

		tmplOpts    []skeletonkit.TemplateOption
		creatorOpts []skeletonkit.CreatorOption

		input string

		wantErr errs
	}{
		"clean":          {"", "example", appInfo{"example", "example.com/example"}, nil, nil, "", errs{false, false, false}},
		"clean-relative": {"", ".", appInfo{"example", "example.com/example"}, nil, nil, "", errs{false, false, false}},

		"overwrite-cancel":          {F(t, "example/main.go", "// not overwritten"), "example", appInfo{"example", "example.com/example"}, nil, nil, "1\n", errs{false, false, false}},
		"overwrite-cancel-relative": {F(t, "main.go", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "1\n", errs{false, false, false}},

		"overwrite-force":          {F(t, "example/main.go", "// not overwritten"), "example", appInfo{"example", "example.com/example"}, nil, nil, "2\n", errs{false, false, false}},
		"overwrite-force-relative": {F(t, "main.go", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "2\n", errs{false, false, false}},

		"overwrite-confirm-yes":          {F(t, "example/main.go", "// not overwritten"), "example", appInfo{"example", "example.com/example"}, nil, nil, "3\ny\n", errs{false, false, false}},
		"overwrite-confirm-yes-relative": {F(t, "main.go", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "3\ny\n", errs{false, false, false}},

		"overwrite-confirm-no":          {F(t, "example/main.go", "// not overwritten"), "example", appInfo{"example", "example.com/example"}, nil, nil, "3\nn\n", errs{false, false, false}},
		"overwrite-confirm-no-relative": {F(t, "main.go", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "3\nn\n", errs{false, false, false}},

		"overwrite-newonly":          {F(t, "example/go.mod", "// not overwritten"), "example", appInfo{"example", "example.com/example"}, nil, nil, "4\n", errs{false, false, false}},
		"overwrite-newonly-relative": {F(t, "go.mod", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "4\n", errs{false, false, false}},

		"prompt-choose-invalidinput": {F(t, "go.mod", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "INVALID\n", errs{false, false, true}},
		"prompt-yesno-invalidinput":  {F(t, "go.mod", "// not overwritten"), ".", appInfo{"example", "example.com/example"}, nil, nil, "3\nINVALID\n", errs{false, false, true}},

		"templateopts-delims": {"", "example", appInfo{"example", "example.com/example"}, T(skeletonkit.TemplateWithDelims("$$", "$$")), nil, "", errs{false, false, false}},
		"templateopts-funcs":  {"", "example", appInfo{"example", "example.com/example"}, T(skeletonkit.TemplateWithFuncs(template.FuncMap{"gomod": func() string { return "DIFFERENT-GOMOD" }})), nil, "", errs{false, false, false}},
		"creatoropts-empty":   {"", "example", appInfo{"example", "example.com/example"}, nil, C(skeletonkit.CreatorWithEmpty(true)), "", errs{false, false, false}},
		"creatoropts-policy":  {F(t, "example/main.go", "// not overwritten"), "example", appInfo{"example", "example.com/example"}, nil, C(skeletonkit.CreatorWithPolicy(skeletonkit.Confirm)), "n\n", errs{false, false, false}},
	}

	if flagUpdate {
		golden.RemoveAll(t, "testdata")
	}

	for name, tt := range cases {
		name, tt := name, tt
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if tt.dirinit != "" {
				golden.DirInit(t, dir, tt.dirinit)
			}

			prompt := &skeletonkit.Prompt{
				Input:     strings.NewReader(tt.input),
				Output:    io.Discard,
				ErrOutput: io.Discard,
			}

			tmpl, err := skeletonkit.ParseTemplate(testTmplFS, "example", "testdata/template", tt.tmplOpts...)
			switch {
			case tt.wantErr.parse && err == nil:
				t.Error("expected error did not occur")
			case !tt.wantErr.parse && err != nil:
				t.Error("unexpected error:", err)
			}

			fsys, err := skeletonkit.ExecuteTemplate(tmpl, tt.info)
			switch {
			case tt.wantErr.exec && err == nil:
				t.Error("expected error did not occur")
			case !tt.wantErr.exec && err != nil:
				t.Error("unexpected error:", err)
			}

			err = skeletonkit.CreateDir(prompt, filepath.Join(dir, tt.root), fsys, tt.creatorOpts...)
			switch {
			case tt.wantErr.create && err == nil:
				t.Error("expected error did not occur")
			case !tt.wantErr.create && err != nil:
				t.Error("unexpected error:", err)
			}

			got := golden.Txtar(t, dir)

			if flagUpdate {
				golden.Update(t, "testdata", name, got)
			}

			if diff := golden.Diff(t, "testdata", name, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
