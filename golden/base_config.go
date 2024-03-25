package golden

import (
	"context"
	"github.com/emirpasic/gods/queues/linkedlistqueue"
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

func (c *BaseConfig) RunPlan() error {
	return c.runDag(dagPlan)
}

func (c *BaseConfig) GetVertices() map[string]interface{} {
	return c.d.GetVertices()
}

func (c *BaseConfig) GetChildren(id string) (map[string]interface{}, error) {
	return c.d.GetChildren(id)
}

func (c *BaseConfig) buildDag(blocks []Block) error {
	return c.d.buildDag(blocks)
}

func (c *BaseConfig) runDag(onReady func(Config, *Dag, *linkedlistqueue.Queue, Block) error) error {
	return c.d.runDag(c, onReady)
}
