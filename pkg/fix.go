package pkg

import (
	"fmt"

	"github.com/Azure/golden"
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type Fix interface {
	golden.ApplyBlock
	GetRuleIds() []string
	// discriminator func
	Fix()
	setRuleIds([]string)
}

var _ golden.Valuable = &BaseFix{}
var _ golden.BaseDecode = &BaseFix{}

type BaseFix struct {
	RuleIds []string `json:"rule_ids" hcl:"rule_ids"`
}

func (bf *BaseFix) Fix() {}

func (bf *BaseFix) BaseDecode(hb *golden.HclBlock, evalContext *hcl.EvalContext) error {
	ruleIdsAttr, ok := hb.Body.Attributes["rule_ids"]
	if !ok {
		return fmt.Errorf("missing required attribute `rule_ids`, every `fix` block must define `rule_ids`")
	}
	ruleIds, diag := ruleIdsAttr.Expr.Value(evalContext)
	if diag.HasErrors() {
		return diag
	}
	var ids []string
	for _, id := range ruleIds.AsValueSlice() {
		ids = append(ids, id.AsString())
	}
	bf.setRuleIds(ids)
	return nil
}

func (bf *BaseFix) ExecuteDuringPlan() error {
	return nil
}

func (bf *BaseFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_ids": golden.ToCtyValue(bf.RuleIds),
	}
}

func (bf *BaseFix) BlockType() string {
	return "fix"
}

func (bf *BaseFix) GetRuleIds() []string {
	return bf.RuleIds
}

func (bf *BaseFix) AddressLength() int { return 3 }

func (bf *BaseFix) CanExecutePrePlan() bool {
	return false
}

func (bf *BaseFix) setRuleIds(ids []string) {
	bf.RuleIds = ids
}
