package pkg

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
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
	SetValues(values map[string]cty.Value)
	SetBaseValues(map[string]cty.Value)
}

type BaseData struct {
	c    *Config
	hb   *hclsyntax.Block
	name string
	id   string
}

func (bd *BaseData) Id() string {
	return bd.id
}

func (bd *BaseData) Name() string {
	return bd.name
}

func (bd *BaseData) BlockType() string {
	return "data"
}

func (bd *BaseData) Parse(b *hclsyntax.Block) error {
	bd.hb = b
	bd.name = b.Labels[1]
	if bd.id == "" {
		bd.id = uuid.NewString()
	}
	return nil
}

func (bd *BaseData) HclSyntaxBlock() *hclsyntax.Block {
	return bd.hb
}

func (bd *BaseData) SetBaseValues(values map[string]cty.Value) {
	values["id"] = cty.StringVal(bd.id)
}

func (bd *BaseData) EvalContext() *hcl.EvalContext {
	return bd.c.EvalContext()
}

func (bd *BaseData) Context() context.Context {
	return bd.c.ctx
}
