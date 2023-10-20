package pkg

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &LocalFile{}

type LocalFile struct {
	*BaseFix
	RuleId  string `json:"rule_id" hcl:"rule_id"`
	Path    string `json:"path" hcl:"path"`
	Content string `json:"content" hcl:"content"`
}

func (lf *LocalFile) Value() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"path":    cty.StringVal(lf.Path),
		"content": cty.StringVal(lf.Content),
	})
}

func (lf *LocalFile) Type() string {
	return "local_file"
}

var _ Fix = &LocalFile{}

func (lf *LocalFile) ApplyFix() error {
	err := afero.WriteFile(FsFactory(), lf.Path, []byte(lf.Content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (lf *LocalFile) Parse(b *hclsyntax.Block) error {
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
