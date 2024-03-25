package golden

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Local = &LocalBlock{}
var _ PlanBlock = &LocalBlock{}
var _ PrePlanBlock = &LocalBlock{}

type Local interface {
	SingleValueBlock
	// discriminator func
	Local()
}

type LocalBlock struct {
	*BaseBlock
	LocalValue cty.Value `hcl:"value"`
}

func (l *LocalBlock) CanExecutePrePlan() bool {
	can := true
	upstreams, _ := l.c.GetAncestors(l.Address())
	for _, i := range upstreams {
		b := i.(Block)
		if !b.CanExecutePrePlan() {
			can = false
		}
	}
	return can
}

func (l *LocalBlock) ExecuteDuringPlan() error {
	return l.parseValue()
}

func (l *LocalBlock) ExecuteBeforePlan() error {
	if l.CanExecutePrePlan() {
		return l.parseValue()
	}
	return nil
}

func (l *LocalBlock) parseValue() error {
	value, diag := l.HclBlock().Body.Attributes["value"].Expr.Value(l.EvalContext())
	if diag.HasErrors() {
		return diag
	}
	l.LocalValue = value
	return nil
}

func (l *LocalBlock) Value() cty.Value {
	return l.LocalValue
}

func (l *LocalBlock) Type() string {
	return ""
}

func (l *LocalBlock) BlockType() string {
	return "local"
}

func (l *LocalBlock) Local() {}

func (l *LocalBlock) AddressLength() int { return 2 }
