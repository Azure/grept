package pkg

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &LocalExecShellFix{}

type LocalExecShellFix struct {
	*BaseBlock
	baseFix
	RuleId        string   `hcl:"rule_id"`
	InlineShebang string   `hcl:"inline_shebang"`
	Inlines       []string `hcl:"inlines" validate:"conflict_with=Script RemoteScript,at_least_one_of=Inlines Script RemoteScript"`
	Script        string   `hcl:"script" validate:"conflict_with=Inlines RemoteScript,at_least_one_of=Inlines Script RemoteScript"`
	RemoteScript  string   `hcl:"remote_script" validate:"conflict_with=Inlines Script,at_least_one_of=Inlines Script RemoteScript"`
}

func (l *LocalExecShellFix) Type() string {
	return "local_exec_shell"
}

func (l *LocalExecShellFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"inlines":       ToCtyValue(l.Inlines),
		"script":        ToCtyValue(l.Script),
		"remote_script": ToCtyValue(l.RemoteScript),
	}
}

func (l *LocalExecShellFix) ApplyFix() error {
	//TODO implement me
	panic("implement me")
}

func (l *LocalExecShellFix) GetRuleId() string {
	return l.RuleId
}
