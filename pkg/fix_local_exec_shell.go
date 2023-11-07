package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/zclconf/go-cty/cty"
	"runtime"
)

var _ Fix = &LocalExecShellFix{}

type LocalExecShellFix struct {
	*BaseBlock
	baseFix
	RuleId        string   `hcl:"rule_id"`
	InlineShebang string   `hcl:"inline_shebang,optional" validate:"required_with=Inlines"`
	Inlines       []string `hcl:"inlines,optional" validate:"conflict_with=Script RemoteScript,at_least_one_of=Inlines Script RemoteScript"`
	Script        string   `hcl:"script,optional" validate:"conflict_with=Inlines RemoteScript,at_least_one_of=Inlines Script RemoteScript"`
	RemoteScript  string   `hcl:"remote_script,optional" validate:"conflict_with=Inlines Script,at_least_one_of=Inlines Script RemoteScript"`
	OnlyOn        []string `hcl:"only_on,optional" validate:"all_string_in_slice=windows linux darwin openbsd netbsd freebsd dragonfly android solaris plan9"`
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

var stopByOnlyOnStub = func() {}

func (l *LocalExecShellFix) ApplyFix() error {
	if len(l.OnlyOn) > 0 && !linq.From(l.OnlyOn).Contains(runtime.GOOS) {
		stopByOnlyOnStub()
		return nil
	}
	panic("implement me")
}

func (l *LocalExecShellFix) GetRuleId() string {
	return l.RuleId
}
