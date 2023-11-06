package pkg

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &DirExistRule{}

type DirExistRule struct {
	*BaseBlock
	baseRule
	Dir string `hcl:"dir"`
}

func (d *DirExistRule) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"dir": ToCtyValue(d.Dir),
	}
}

func (d *DirExistRule) Type() string {
	return "dir_exist"
}

func (d *DirExistRule) Check() (checkError error, runtimeError error) {
	fs := FsFactory()
	exists, err := afero.Exists(fs, d.Dir)
	if err != nil {
		runtimeError = err
		return
	}
	if !exists {
		checkError = fmt.Errorf("directory does not exist: %s", d.Dir)
	}
	return
}
