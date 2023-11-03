package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Data interface {
	Load() error
	Type() string
	BlockType() string
	Name() string
	Eval(*hclsyntax.Block) error
	HclSyntaxBlock() *hclsyntax.Block
	Id() string
	Values() map[string]cty.Value
	BaseValues() map[string]cty.Value
}

type baseData struct{}

func (bd baseData) BlockType() string {
	return "data"
}
