package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Fix interface {
	ApplyFix() error
	GetRule() string
	Parse(b *hclsyntax.Block) error
}

type BaseFix struct {
	Rule string
	ctx  *hcl.EvalContext
}

func (bf *BaseFix) GetRule() string {
	return bf.Rule
}

func (bf *BaseFix) Parse(b *hclsyntax.Block) (err error) {
	if len(b.Labels) != 2 {
		return fmt.Errorf("invalid labels for %s %s, expect labels with length 2", b.Type, concatLabels(b.Labels))
	}
	bf.Rule, err = readRequiredStringAttribute(b, "rule", bf.ctx)
	if err != nil {
		return fmt.Errorf("unrecognized rule: %s", bf.Rule)
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
