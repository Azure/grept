package pkg

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &DirExistRule{}

type DirExistRule struct {
	*BaseBlock
	*BaseRule
	Dir         string `hcl:"dir"`
	FailOnExist bool   `hcl:"fail_on_exist,optional"`
}

func (d *DirExistRule) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"dir":            ToCtyValue(d.Dir),
		"faile_on_exist": ToCtyValue(d.FailOnExist),
	}
}

func (d *DirExistRule) Type() string {
	return "dir_exist"
}

func (d *DirExistRule) ExecuteDuringPlan() error {
	fs := FsFactory()
	exists, err := afero.Exists(fs, d.Dir)
	if err != nil {
		return err
	}
	if d.FailOnExist && exists {
		d.setCheckError(fmt.Errorf("directory exists: %s", d.Dir))
		return nil
	}
	if !d.FailOnExist && !exists {
		d.setCheckError(fmt.Errorf("directory does not exist: %s", d.Dir))
		return nil
	}
	return nil
}
