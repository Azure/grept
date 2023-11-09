package pkg

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/mcuadros/go-defaults"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

type block interface {
	Id() string
	Name() string
	Type() string
	BlockType() string
	HclSyntaxBlock() *hclsyntax.Block
	EvalContext() *hcl.EvalContext
	Values() map[string]cty.Value
	BaseValues() map[string]cty.Value
	parseBase(*hclsyntax.Block) error
}

func blockToString(f block) string {
	marshal, _ := json.Marshal(f)
	return string(marshal)
}

func eval(b block) error {
	hb := b.HclSyntaxBlock()
	err := b.parseBase(hb)
	if err != nil {
		return err
	}
	defaults.SetDefaults(b)
	diag := gohcl.DecodeBody(hb.Body, b.EvalContext(), b)
	if diag.HasErrors() {
		return diag
	}
	return nil
}

func Values[T block](blocks []T) cty.Value {
	if len(blocks) == 0 {
		return cty.EmptyObjectVal
	}
	res := map[string]cty.Value{}
	valuesMap := map[string]map[string]cty.Value{}

	for _, b := range blocks {
		values := valuesMap[b.Type()]
		if values == nil {
			values = map[string]cty.Value{}
			valuesMap[b.Type()] = values
		}
		blockValues := map[string]cty.Value{}
		baseCtyValues := b.BaseValues()
		ctyValues := b.Values()
		for k, v := range ctyValues {
			blockValues[k] = v
		}
		for k, v := range baseCtyValues {
			blockValues[k] = v
		}
		values[b.Name()] = cty.ObjectVal(blockValues)
	}
	for t, m := range valuesMap {
		res[t] = cty.MapVal(m)
	}
	return cty.ObjectVal(res)
}

func concatLabels(labels []string) string {
	sb := strings.Builder{}
	for i, l := range labels {
		sb.WriteString(l)
		if i != len(labels)-1 {
			sb.WriteString(".")
		}
	}
	return sb.String()
}

func refresh(b block) {
	_ = eval(b)
}

func blockAddress(b block) string {
	sb := strings.Builder{}
	sb.WriteString(b.BlockType())
	sb.WriteString(".")
	if t := b.Type(); t != "" {
		sb.WriteString(t)
		sb.WriteString(".")
	}
	sb.WriteString(b.Name())
	return sb.String()
}

type BaseBlock struct {
	c    *Config
	hb   *hclsyntax.Block
	name string
	id   string
}

func (bb *BaseBlock) Id() string {
	return bb.id
}

func (bb *BaseBlock) Name() string {
	return bb.name
}

func (bb *BaseBlock) HclSyntaxBlock() *hclsyntax.Block {
	if bb.hb == nil {
		return new(hclsyntax.Block)
	}
	return bb.hb
}

func (bb *BaseBlock) BaseValues() map[string]cty.Value {
	return map[string]cty.Value{
		"id": cty.StringVal(bb.id),
	}
}

func (bb *BaseBlock) EvalContext() *hcl.EvalContext {
	if bb.c == nil {
		return new(hcl.EvalContext)
	}
	return bb.c.EvalContext()
}

func (bb *BaseBlock) Context() context.Context {
	if bb.c == nil {
		return context.TODO()
	}
	return bb.c.ctx
}

func (bb *BaseBlock) parseBase(b *hclsyntax.Block) error {
	bb.hb = b
	bb.name = b.Labels[1]
	if bb.id == "" {
		bb.id = uuid.NewString()
	}
	return nil
}
