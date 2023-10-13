package pkg

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Data interface {
	Load(context.Context) error
	Type() string
	Name() string
	Value() cty.Value
	Parse(*hclsyntax.Block) error
	Id() string
}

type BaseData struct {
	c    *Config
	name string
	id   string
}

func (bd *BaseData) Id() string {
	return bd.id
}

func (bd *BaseData) Name() string {
	return bd.name
}

func (bd *BaseData) Parse(b *hclsyntax.Block) error {
	bd.name = b.Labels[1]
	bd.id = uuid.NewString()
	return nil
}

func (bd *BaseData) EvalContext() *hcl.EvalContext {
	return bd.c.EvalContext()
}
