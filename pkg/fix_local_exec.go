package pkg

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &LocalExecFix{}

type LocalExecFix struct {
	*BaseBlock
	baseFix
	RuleId       string   `hcl:"rule_id"`
	Inlines      []string `hcl:"inlines" validate:"conflict_with=Script RemoteScript"`
	Script       string   `hcl:"script" validate:"conflict_with=Inlines RemoteScript"`
	RemoteScript string   `hcl:"remote_script" validate:"conflict_with=Inlines Script"`
}

func (l *LocalExecFix) Type() string {
	return "local_exec"
}

func (l *LocalExecFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"inlines":       ToCtyValue(l.Inlines),
		"script":        ToCtyValue(l.Script),
		"remote_script": ToCtyValue(l.RemoteScript),
	}
}

func (l *LocalExecFix) ApplyFix() error {
	//TODO implement me
	panic("implement me")
}

func (l *LocalExecFix) GetRuleId() string {
	return l.RuleId
}
