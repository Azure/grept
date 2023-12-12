package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type hclBlock struct {
	*hclsyntax.Block
	*forEach
}

func newHclBlock(hb *hclsyntax.Block, each *forEach) *hclBlock {
	return &hclBlock{
		Block:   hb,
		forEach: each,
	}
}

type forEach struct {
	key   cty.Value
	value cty.Value
}
