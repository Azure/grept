package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Fix interface {
	ApplyFix() error
	GetRule() string
	Validate() error
	Parse(b *hclsyntax.Block) error
}

type BaseFix struct {
	Rule string `hcl:"rule,attr"`
	ctx  *hcl.EvalContext
}

func (bf *BaseFix) GetRule() string {
	return bf.Rule
}

func (bf *BaseFix) Parse(b *hclsyntax.Block) (err error) {
	bf.Rule, err = readRequiredStringAttribute(b, "rule", bf.ctx)
	return
}

func (bf *BaseFix) Validate() error {
	if bf.Rule == "" {
		return fmt.Errorf("`rule` is required")
	}
	return nil
}

var FixFactories = map[string]func(*hcl.EvalContext) Fix{}

func registerFix() {
	FixFactories["local_file"] = func(ctx *hcl.EvalContext) Fix {
		return &LocalFile{
			BaseFix: &BaseFix{
				ctx: ctx,
			},
		}
	}
}
