package pkg

import (
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Rule interface {
	Check() (checkError error, runtimeError error)
	Type() string
	BlockType() string
	Name() string
	Id() string
	Eval(*hclsyntax.Block) error
	HclSyntaxBlock() *hclsyntax.Block
	SetValues(values map[string]cty.Value)
	SetBaseValues(map[string]cty.Value)
}

type BaseRule struct {
	c    *Config
	hb   *hclsyntax.Block
	name string
	id   string
}

func (br *BaseRule) Parse(b *hclsyntax.Block) error {
	br.hb = b
	br.name = b.Labels[1]
	if br.id == "" {
		br.id = uuid.NewString()
	}
	return nil
}

func (br *BaseRule) HclSyntaxBlock() *hclsyntax.Block {
	return br.hb
}

func (br *BaseRule) Id() string {
	return br.id
}

func (br *BaseRule) Name() string {
	return br.name
}

func (br *BaseRule) BlockType() string {
	return "rule"
}

func (br *BaseRule) SetBaseValues(values map[string]cty.Value) {
	values["id"] = cty.StringVal(br.id)
}

func (br *BaseRule) EvalContext() *hcl.EvalContext {
	return br.c.EvalContext()
}
