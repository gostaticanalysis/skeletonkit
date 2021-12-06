package skeletonkit

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type overwritePolicy int

const (
	promptAction overwritePolicy = iota
	Cancel
	ForceOverwrite
	Confirm
	NewOnly
)

// Creator is the representation of a file and directory writer.
type Creator struct {
	includeEmpty bool
	policy       overwritePolicy
}

// CreatorOption is an option for a creator.
// It can decorate a creator.
type CreatorOption func(*Creator)

// CreatorWithEmpty sets if a creator must write empty files.
func CreatorWithEmpty(include bool) CreatorOption {
	return func(c *Creator) {
		c.includeEmpty = include
	}
}

// CreatorWithPolicy sets the overwriting policy of a creator.
func CreatorWithPolicy(policy overwritePolicy) CreatorOption {
	return func(c *Creator) {
		c.policy = policy
	}
}

// CreateDir creates files and directories which structure is the same with the given file system.
// The path of created root directory become the parameter root.
func CreateDir(prompt *Prompt, root string, fsys fs.FS, options ...CreatorOption) error {
	creator := &Creator{}
	for _, opt := range options {
		opt(creator)
	}

	var err error
	if creator.policy == promptAction {
		creator.policy, err = choosePolicy(prompt, root)
		if err != nil {
			return err
		}
	}

	if creator.policy == ForceOverwrite {
		if err := removeDir(root); err != nil {
			return err
		}
	}

	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) (rerr error) {
		if err != nil {
			return err
		}

		// directory would create with a file
		if d.IsDir() {
			return nil
		}

		dstPath := filepath.Join(root, filepath.FromSlash(path))

		src, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		fi, err := src.Stat()
		if err != nil {
			return err
		}

		if !creator.includeEmpty && fi.Size() == 0 {
			return nil
		}

		err = os.MkdirAll(filepath.Dir(dstPath), 0700)
		if err != nil {
			return err
		}

		dst, err := create(prompt, dstPath, creator.policy)
		if err != nil {
			return err
		}
		defer func() {
			if err := dst.Close(); err != nil && rerr == nil {
				rerr = err
			}
		}()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("CreateDir: %w", err)
	}
	return nil
}

func choosePolicy(prompt *Prompt, dir string) (overwritePolicy, error) {
	exist, err := isExist(dir)
	if err != nil {
		return 0, err
	}

	if !exist {
		return Cancel, nil
	}

	desc := fmt.Sprintf("%s is already exist, overwrite?", dir)
	opts := []string{
		Cancel:         "No (Exit)",
		ForceOverwrite: "Remove and create new directory",
		Confirm:        "Overwrite existing files with confirmation",
		NewOnly:        "Create new files only",
	}
	n, err := prompt.Choose(desc, opts, ">")
	if err != nil {
		return 0, err
	}

	return overwritePolicy(n), nil
}

func isExist(dir string) (bool, error) {
	d, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer d.Close()

	_, err = d.Readdirnames(1)
	if err != nil {
		if err == io.EOF {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func removeDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}

	return nil
}

func create(prompt *Prompt, path string, policy overwritePolicy) (io.WriteCloser, error) {
	var nopWriter = struct {
		io.Writer
		io.Closer
	}{io.Discard, io.NopCloser(nil)}

	exist, err := isExist(path)
	if err != nil {
		return nil, err
	}

	if !exist {
		return os.Create(path)
	}

	if policy != Confirm {
		return nopWriter, nil
	}

	desc := fmt.Sprintf("%s is already exist, overwrite?", path)
	yesno, err := prompt.YesNo(desc, false, '>')
	if err != nil {
		return nil, err
	}

	if !yesno {
		return nopWriter, nil
	}

	return os.Create(path)
}
