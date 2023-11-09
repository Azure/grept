package pkg

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &FileExistRule{}

type FileExistRule struct {
	*BaseBlock
	baseRule
	Glob       string `hcl:"glob"`
	MatchFiles []string
}

func (f *FileExistRule) Type() string {
	return "file_exist"
}

func (f *FileExistRule) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"glob":        ToCtyValue(f.Glob),
		"match_files": ToCtyValue(f.MatchFiles),
	}
}

func (f *FileExistRule) Check() (checkError error, runtimeError error) {
	fs := FsFactory()
	finds, err := afero.Glob(fs, f.Glob)
	if err != nil {
		return nil, fmt.Errorf("error on glob files %s, rule.%s.%s %s", f.Glob, f.Type(), f.Name(), f.HclSyntaxBlock().Range().String())
	}
	if len(finds) == 0 {
		return fmt.Errorf("no match on glob %s, rule.%s.%s %s", f.Glob, f.Type(), f.Name(), f.HclSyntaxBlock().Range().String()), nil
	}
	f.MatchFiles = finds
	return nil, nil
}
