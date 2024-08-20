package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/spf13/afero"
)

var _ Rule = &FileExistRule{}

type FileExistRule struct {
	*golden.BaseBlock
	*BaseRule
	Glob       string `hcl:"glob"`
	MatchFiles []string
}

func (f *FileExistRule) Type() string {
	return "file_exist"
}

func (f *FileExistRule) ExecuteDuringPlan() error {
	fs := FsFactory()
	finds, err := afero.Glob(fs, f.Glob)
	if err != nil {
		return fmt.Errorf("error on glob files %s, %s", f.Glob, f.Address())
	}
	if len(finds) == 0 {
		f.setCheckError(fmt.Errorf("no match on glob %s, %s", f.Glob, f.Address()))
		return nil
	}
	f.MatchFiles = finds
	return nil
}
