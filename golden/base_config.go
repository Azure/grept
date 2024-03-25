package golden

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/lonegunmanb/hclfuncs"
	"github.com/zclconf/go-cty/cty"
)

type BaseConfig struct {
	ctx     context.Context
	basedir string
	d       *Dag
}

func (c *BaseConfig) Context() context.Context {
	return c.ctx
}

func (c *BaseConfig) dag() *Dag {
	return c.d
}

func (c *BaseConfig) blocksByTypes() map[string][]Block {
	r := make(map[string][]Block)
	for _, b := range blocks(c) {
		bt := b.BlockType()
		r[bt] = append(r[bt], b)
	}
	return r
}

func (c *BaseConfig) EvalContext() *hcl.EvalContext {
	ctx := hcl.EvalContext{
		Functions: hclfuncs.Functions(c.basedir),
		Variables: make(map[string]cty.Value),
	}
	for bt, bs := range c.blocksByTypes() {
		sample := bs[0]
		if _, ok := sample.(SingleValueBlock); ok {
			ctx.Variables[bt] = SingleValues(castBlock[SingleValueBlock](bs))
			continue
		}
		ctx.Variables[bt] = Values(bs)
	}
	return &ctx
}

func NewBasicConfig(basedir string, ctx context.Context) *BaseConfig {
	if ctx == nil {
		ctx = context.Background()
	}
	c := &BaseConfig{
		basedir: basedir,
		ctx:     ctx,
		d:       newDag(),
	}
	return c
}

func (c *BaseConfig) RunPrePlan() error {
	return c.runDag(prePlan)
}

func (c *BaseConfig) RunPlan() error {
	return c.runDag(dagPlan)
}

func (c *BaseConfig) GetVertices() map[string]interface{} {
	return c.d.GetVertices()
}

func (c *BaseConfig) GetAncestors(id string) (map[string]interface{}, error) {
	return c.d.GetAncestors(id)
}

func (c *BaseConfig) GetChildren(id string) (map[string]interface{}, error) {
	return c.d.GetChildren(id)
}

func (c *BaseConfig) buildDag(blocks []Block) error {
	return c.d.buildDag(blocks)
}

func (c *BaseConfig) runDag(onReady func(Block) error) error {
	return c.d.runDag(c, onReady)
}

func (c *BaseConfig) expandBlock(b Block) ([]Block, error) {
	var expandedBlocks []Block
	attr, ok := b.HclBlock().Body.Attributes["for_each"]
	if !ok || b.getForEach() != nil {
		return nil, nil
	}
	forEachValue, diag := attr.Expr.Value(c.EvalContext())
	if diag.HasErrors() {
		return nil, diag
	}
	if !forEachValue.CanIterateElements() {
		return nil, fmt.Errorf("invalid `for_each`, except set or map: %s", attr.Range().String())
	}
	address := b.Address()
	upstreams, err := c.d.GetAncestors(address)
	if err != nil {
		return nil, err
	}
	downstreams, err := c.d.GetChildren(address)
	if err != nil {
		return nil, err
	}
	iterator := forEachValue.ElementIterator()
	for iterator.Next() {
		key, value := iterator.Element()
		newBlock := NewHclBlock(b.HclBlock().Block, &forEach{key: key, value: value})
		nb, err := wrapBlock(c, newBlock)
		if err != nil {
			return nil, err
		}
		nb.markExpanded()
		expandedAddress := blockAddress(newBlock)
		expandedBlocks = append(expandedBlocks, nb)
		err = c.d.AddVertexByID(expandedAddress, nb)
		if err != nil {
			return nil, err
		}
		for upstreamAddress := range upstreams {
			err := c.d.addEdge(upstreamAddress, expandedAddress)
			if err != nil {
				return nil, err
			}
		}
		for downstreamAddress := range downstreams {
			err := c.d.addEdge(expandedAddress, downstreamAddress)
			if err != nil {
				return nil, err
			}
		}
	}
	b.markExpanded()
	return expandedBlocks, c.d.DeleteVertex(address)
}
