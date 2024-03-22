package golden

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type HclBlock struct {
	*hclsyntax.Block
	*forEach
}

func NewHclBlock(hb *hclsyntax.Block, each *forEach) *HclBlock {
	return &HclBlock{
		Block:   hb,
		forEach: each,
	}
}

type forEach struct {
	key   cty.Value
	value cty.Value
}

func AsHclBlocks(bs hclsyntax.Blocks) []*HclBlock {
	var blocks []*HclBlock
	for _, b := range bs {
		var bs = readRawHclBlock(b)
		for _, hb := range bs {
			blocks = append(blocks, NewHclBlock(hb, nil))
		}
	}
	return blocks
}

func readRawHclBlock(b *hclsyntax.Block) []*hclsyntax.Block {
	if b.Type != "locals" {
		return []*hclsyntax.Block{b}
	}
	var newBlocks []*hclsyntax.Block
	for _, attr := range b.Body.Attributes {
		newBlocks = append(newBlocks, &hclsyntax.Block{
			Type:   "local",
			Labels: []string{"", attr.Name},
			Body: &hclsyntax.Body{
				Attributes: map[string]*hclsyntax.Attribute{
					"value": {
						Name:        "value",
						Expr:        attr.Expr,
						SrcRange:    attr.SrcRange,
						NameRange:   attr.NameRange,
						EqualsRange: attr.EqualsRange,
					},
				},
				SrcRange: attr.NameRange,
				EndRange: attr.SrcRange,
			},
		})
	}
	return newBlocks
}
