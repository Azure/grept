package pkg

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &MustBeTrueRule{}

type MustBeTrueRule struct {
	*BaseBlock
	baseRule
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

func (m *MustBeTrueRule) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"condition":     ToCtyValue(m.Condition),
		"error_message": ToCtyValue(m.ErrorMessage),
	}
}
