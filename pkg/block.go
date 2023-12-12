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
	"sync"
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
	setOperator(o *BlocksOperator)
	notifyOnExecuted(b block, success bool)
	forEachDefined() bool
	setOnReady(func(*Config, block))
	getDownstreams() []block
}

func blockToString(f block) string {
	marshal, _ := json.Marshal(f)
	return string(marshal)
}

func decode(b block) error {
	defaults.SetDefaults(b)
	hb := b.HclSyntaxBlock()
	diag := gohcl.DecodeBody(hb.Body, b.EvalContext(), b)
	if diag.HasErrors() {
		return diag
	}
	return nil
}

func LocalsValues(blocks []Local) cty.Value {
	if len(blocks) == 0 {
		return cty.EmptyObjectVal
	}
	res := map[string]cty.Value{}
	for _, b := range blocks {
		for _, v := range b.Values() {
			res[b.Name()] = v
		}
	}
	return cty.ObjectVal(res)
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
		if l == "" {
			continue
		}
		sb.WriteString(l)
		if i != len(labels)-1 {
			sb.WriteString(".")
		}
	}
	return sb.String()
}

func blockAddress(b *hclsyntax.Block) string {
	sb := strings.Builder{}
	sb.WriteString(b.Type)
	sb.WriteString(".")
	sb.WriteString(concatLabels(b.Labels))
	return sb.String()
}

type BaseBlock struct {
	c            *Config
	hb           *hclsyntax.Block
	name         string
	id           string
	operator     *BlocksOperator
	blockAddress string
	mu           sync.Mutex
	onReady      func(*Config, block)
}

func newBaseBlock(c *Config, hb *hclsyntax.Block) *BaseBlock {
	bb := &BaseBlock{
		c:            c,
		hb:           hb,
		blockAddress: blockAddress(hb),
		name:         hb.Labels[1],
		id:           uuid.NewString(),
	}
	return bb
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

func (bb *BaseBlock) setOperator(o *BlocksOperator) {
	bb.operator = o
}

func (bb *BaseBlock) notifyOnExecuted(b block, success bool) {
	bb.mu.Lock()
	pendingUpstreams := bb.c.dag.pendingUpstreams[bb.blockAddress]
	pendingUpstreams.Remove(blockAddress(b.HclSyntaxBlock()))
	bb.mu.Unlock()
	self, _ := bb.c.dag.GetVertex(bb.blockAddress)
	selfBlock := self.(block)
	if !success {
		bb.c.notifyOnExecuted(selfBlock, false)
	}
	if pendingUpstreams.Empty() {
		go func() {
			if bb.onReady != nil {
				bb.onReady(bb.c, selfBlock)
			}
		}()
	}
}

func (bb *BaseBlock) forEachDefined() bool {
	_, forEach := bb.HclSyntaxBlock().Body.Attributes["for_each"]
	return forEach
}

func (bb *BaseBlock) setOnReady(next func(*Config, block)) {
	bb.onReady = next
}

func (bb *BaseBlock) getDownstreams() []block {
	var blocks []block
	children, _ := bb.c.dag.GetChildren(bb.blockAddress)
	for _, c := range children {
		blocks = append(blocks, c.(block))
	}
	return blocks
}

func plan(c *Config, b block) error {
	self, _ := c.dag.GetVertex(blockAddress(b.HclSyntaxBlock()))
	return c.planBlock(self.(block))
}

func prepare(c *Config, b block) error {
	l, ok := b.(*LocalBlock)
	if !ok {
		return nil
	}
	value, diag := l.HclSyntaxBlock().Body.Attributes["value"].Expr.Value(c.EvalContext())
	if !diag.HasErrors() {
		l.Value = value
		return nil
	}
	return diag
}
