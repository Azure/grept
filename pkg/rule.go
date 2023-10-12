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
	ctx  *hcl.EvalContext
	name string
	id   string
}

func (r *BaseRule) Parse(b *hclsyntax.Block) error {
	r.name = b.Labels[1]
	r.id = uuid.NewString()
	return nil
}

func (r *BaseRule) Id() string {
	return r.id
}

func (r *BaseRule) Name() string {
	return r.name
}

func (r *BaseRule) BaseValue() map[string]cty.Value {
	return map[string]cty.Value{
		"id": cty.StringVal(r.id),
	}
}
