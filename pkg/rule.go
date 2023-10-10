package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func init() {
	registerRule()
	registerFix()
}

type Rule interface {
	Check() error
	Parse(*hclsyntax.Block) error
}

type BaseRule struct {
	ctx *hcl.EvalContext
}

func (r *BaseRule) Parse(*hclsyntax.Block) error {
	return nil
}

var RuleFactories = map[string]func(*hcl.EvalContext) Rule{}

func registerRule() {
	RuleFactories["file_hash"] = func(ctx *hcl.EvalContext) Rule {
		return &FileHashRule{
			BaseRule: &BaseRule{
				ctx: ctx,
			},
		}
	}
}
