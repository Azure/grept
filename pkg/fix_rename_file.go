package pkg

import (
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &RenameFileFix{}

type RenameFileFix struct {
	*BaseBlock
	baseFix
	RuleIds []string `json:"rule_ids" hcl:"rule_ids"`
	OldName string   `json:"old_name" hcl:"old_name"`
	NewName string   `json:"new_name" hcl:"new_name"`
}

func (rf *RenameFileFix) GetRuleIds() []string {
	return rf.RuleIds
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
		"rule_id":  ToCtyValue(rf.RuleIds),
		"old_name": ToCtyValue(rf.OldName),
		"new_name": ToCtyValue(rf.NewName),
	}
}
