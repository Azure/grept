package pkg

import (
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &RmLocalFile{}

type RmLocalFile struct {
	*BaseFix
	Paths []string `hcl:"paths"`
}

func (r *RmLocalFile) Type() string {
	return "rm_local_file"
}

func (r *RmLocalFile) ApplyFix() error {
	fs := FsFactory()
	var err error
	for _, path := range r.Paths {
		removeErr := fs.Remove(path)
		if removeErr != nil {
			err = multierror.Append(err, removeErr)
		}
	}
	return err
}

func (r *RmLocalFile) Eval(b *hclsyntax.Block) error {
	err := r.BaseFix.Parse(b)
	if err != nil {
		return err
	}
	diag := gohcl.DecodeBody(b.Body, r.EvalContext(), r)
	if diag.HasErrors() {
		return diag
	}
	return err
}

func (r *RmLocalFile) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"paths": ToCtyValue(r.Paths),
	}
}
