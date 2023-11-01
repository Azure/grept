package pkg

import (
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &LocalFile{}

type LocalFile struct {
	*BaseFix
	RuleId  string   `json:"rule_id" hcl:"rule_id"`
	Paths   []string `json:"paths" hcl:"paths"`
	Content string   `json:"content" hcl:"content"`
}

func (lf *LocalFile) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_id": ToCtyValue(lf.RuleId),
		"paths":   ToCtyValue(lf.Paths),
		"content": ToCtyValue(lf.Content),
	}
}

func (lf *LocalFile) Type() string {
	return "local_file"
}

var _ Fix = &LocalFile{}

func (lf *LocalFile) ApplyFix() error {
	var err error
	for _, path := range lf.Paths {
		writeErr := afero.WriteFile(FsFactory(), path, []byte(lf.Content), 0644)
		if writeErr != nil {
			err = multierror.Append(err, writeErr)
		}
	}

	return err
}

func (lf *LocalFile) Eval(b *hclsyntax.Block) error {
	err := lf.BaseFix.Parse(b)
	if err != nil {
		return err
	}
	diag := gohcl.DecodeBody(b.Body, lf.EvalContext(), lf)
	if diag.HasErrors() {
		return diag
	}
	return err
}
