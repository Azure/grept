package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Fix interface {
	Type() string
	BlockType() string
	Name() string
	ApplyFix() error
	GetRuleId() string
	Eval(b *hclsyntax.Block) error
	HclSyntaxBlock() *hclsyntax.Block
	Values() map[string]cty.Value
	BaseValues() map[string]cty.Value
}

func FixToString(f Fix) string {
	marshal, _ := json.Marshal(f)
	return string(marshal)
}

type BaseFix struct {
	name   string `json:"name" hcl:"name"`
	RuleId string `json:"rule_id" hcl:"rule_id"`
	c      *Config
	hb     *hclsyntax.Block
	id     string
}

func (bf *BaseFix) GetRuleId() string {
	return bf.RuleId
}

func (bf *BaseFix) Parse(b *hclsyntax.Block) (err error) {
	bf.hb = b
	bf.RuleId, err = readRequiredStringAttribute(b, "rule_id", bf.EvalContext())
	if err != nil {
		return fmt.Errorf("cannot parse rule: %s, %s", b.Range().String(), err.Error())
	}
	bf.name = b.Labels[1]
	if bf.id == "" {
		bf.id = uuid.NewString()
	}
	return nil
}

func (bf *BaseFix) HclSyntaxBlock() *hclsyntax.Block {
	return bf.hb
}

func (bf *BaseFix) Name() string {
	return bf.name
}

func (bf *BaseFix) BlockType() string {
	return "fix"
}

func (bf *BaseFix) EvalContext() *hcl.EvalContext {
	return bf.c.EvalContext()
}

func (bf *BaseFix) Context() context.Context {
	return bf.Context()
}

func (bf *BaseFix) BaseValues() map[string]cty.Value {
	return map[string]cty.Value{
		"id": cty.StringVal(bf.id),
	}
}
