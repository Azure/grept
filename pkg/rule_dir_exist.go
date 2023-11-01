package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &DirExistRule{}

type DirExistRule struct {
	*BaseRule
	Dir string `hcl:"dir"`
}

func (d *DirExistRule) Eval(b *hclsyntax.Block) error {
	err := d.BaseRule.Parse(b)
	if err != nil {
		return err
	}
	diag := gohcl.DecodeBody(b.Body, d.EvalContext(), d)
	if diag.HasErrors() {
		return diag
	}
	return nil
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
