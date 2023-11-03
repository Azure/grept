package pkg

import (
	"encoding/json"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Fix interface {
	Type() string
	BlockType() string
	Name() string
	ApplyFix() error
	GetRuleId() string
	Eval(b *hclsyntax.Block) error
	HclSyntaxBlock() *hclsyntax.Block
	Values() map[string]cty.Value
	BaseValues() map[string]cty.Value
}

func blockToString(f block) string {
	marshal, _ := json.Marshal(f)
	return string(marshal)
}

type baseFix struct{}

func (bf baseFix) BlockType() string {
	return "fix"
}
