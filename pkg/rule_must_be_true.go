package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &MustBeTrueRule{}

type MustBeTrueRule struct {
	*BaseRule
	Condition    bool   `hcl:"condition"`
	ErrorMessage string `hcl:"error_message,optional"`
}

func (m *MustBeTrueRule) Check() (checkError error, runtimeError error) {
	if !m.Condition {
		checkError = fmt.Errorf("assertion failed: %s", m.ErrorMessage)
	}
	return
}

func (m *MustBeTrueRule) Type() string {
	return "must_be_true"
}

func (m *MustBeTrueRule) Eval(b *hclsyntax.Block) error {
	err := m.BaseRule.Parse(b)
	if err != nil {
		return err
	}
	diag := gohcl.DecodeBody(b.Body, m.EvalContext(), m)
	if diag.HasErrors() {
		return diag
	}
	return nil
}

func (m *MustBeTrueRule) SetValues(value map[string]cty.Value) {
	value["condition"] = cty.BoolVal(m.Condition)
	value["error_message"] = cty.StringVal(m.ErrorMessage)
}
