package pkg

import (
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Rule interface {
	Check() error
	Type() string
	Name() string
	Id() string
	Parse(*hclsyntax.Block) error
	Value() cty.Value
}

type BaseRule struct {
	c    *Config
	name string
	id   string
}

func (br *BaseRule) Parse(b *hclsyntax.Block) error {
	br.name = b.Labels[1]
	br.id = uuid.NewString()
	return nil
}

func (br *BaseRule) Id() string {
	return br.id
}

func (br *BaseRule) Name() string {
	return br.name
}

func (br *BaseRule) BaseValue() map[string]cty.Value {
	return map[string]cty.Value{
		"id": cty.StringVal(br.id),
	}
}

func (br *BaseRule) EvalContext() *hcl.EvalContext {
	return br.c.EvalContext()
}
