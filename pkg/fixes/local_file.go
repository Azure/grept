package fixes

import (
	"github.com/spf13/afero"
)

type LocalFile struct {
	fs      afero.Fs
	Rule    string `hcl:"rule,attr"`
	Path    string `hcl:"path,attr"`
	Content string `hcl:"content,attr"`
}

func (lf *LocalFile) ApplyFix() error {
	err := afero.WriteFile(lf.fs, lf.Path, []byte(lf.Content), 0644)
	if err != nil {
		return err
	}

	return nil
}
