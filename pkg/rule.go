package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Rule interface {
	Check() (checkError error, runtimeError error)
	Type() string
	BlockType() string
	Name() string
	Id() string
	Eval(*hclsyntax.Block) error
	HclSyntaxBlock() *hclsyntax.Block
	Values() map[string]cty.Value
	BaseValues() map[string]cty.Value
}

type baseRule struct{}

func (br baseRule) BlockType() string {
	return "rule"
}
