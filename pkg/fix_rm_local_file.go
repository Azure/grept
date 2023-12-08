package pkg

import (
	"github.com/hashicorp/go-multierror"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &RmLocalFileFix{}

type RmLocalFileFix struct {
	*BaseBlock
	baseFix
	RuleIds []string `hcl:"rule_ids" json:"rule_ids"`
	Paths   []string `hcl:"paths" json:"paths"`
}

func (r *RmLocalFileFix) GetRuleIds() []string {
	return r.RuleIds
}

func (r *RmLocalFileFix) Type() string {
	return "rm_local_file"
}

func (r *RmLocalFileFix) Apply() error {
	fs := FsFactory()
	var err error
	for _, path := range r.Paths {
		removeErr := fs.RemoveAll(path)
		if removeErr != nil {
			err = multierror.Append(err, removeErr)
		}
	}
	return err
}

func (r *RmLocalFileFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"paths": ToCtyValue(r.Paths),
	}
}
