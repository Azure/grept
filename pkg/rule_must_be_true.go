package pkg

import (
	"fmt"
	"github.com/zclconf/go-cty/cty"
)

var _ Rule = &MustBeTrueRule{}

type MustBeTrueRule struct {
	*BaseBlock
	BaseRule
	Condition    bool   `hcl:"condition"`
	ErrorMessage string `hcl:"error_message,optional"`
}

func (m *MustBeTrueRule) Execute() error {
	if !m.Condition {
		logCheckError(m, fmt.Errorf("assertion failed: %s", m.ErrorMessage))
	}
	return nil
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
