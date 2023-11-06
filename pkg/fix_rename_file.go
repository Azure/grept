package pkg

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &RenameFileFix{}

type RenameFileFix struct {
	*BaseBlock
	baseFix
	RuleId  string `json:"rule_id" hcl:"rule_id"`
	OldName string `json:"old_name" hcl:"old_name"`
	NewName string `json:"new_name" hcl:"new_name"`
}

func (rf *RenameFileFix) GetRuleId() string {
	return rf.RuleId
}

func (rf *RenameFileFix) Type() string {
	return "rename_file"
}

func (rf *RenameFileFix) ApplyFix() error {
	fs := FsFactory()
	return fs.Rename(rf.OldName, rf.NewName)
}

func (rf *RenameFileFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_id":  ToCtyValue(rf.RuleId),
		"old_name": ToCtyValue(rf.OldName),
		"new_name": ToCtyValue(rf.NewName),
	}
}
