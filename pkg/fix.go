package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Fix interface {
	Type() string
	Name() string
	ApplyFix() error
	GetRuleId() string
	Parse(b *hclsyntax.Block) error
}

type BaseFix struct {
	name   string
	RuleId string
	ctx    *hcl.EvalContext
}

func (bf *BaseFix) GetRuleId() string {
	return bf.RuleId
}

func (bf *BaseFix) Parse(b *hclsyntax.Block) (err error) {
	bf.RuleId, err = readRequiredStringAttribute(b, "rule_id", bf.ctx)
	if err != nil {
		return fmt.Errorf("cannot parse rule: %s, %s", b.Range().String(), err.Error())
	}
	bf.name = b.Labels[1]
	return nil
}

func (bf *BaseFix) Name() string {
	return bf.name
}
