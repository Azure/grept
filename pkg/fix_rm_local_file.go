package pkg

import (
	"github.com/hashicorp/go-multierror"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &RmLocalFileFix{}

type RmLocalFileFix struct {
	*BaseBlock
	baseFix
	RuleId string   `hcl:"rule_id"`
	Paths  []string `hcl:"paths"`
}

func (r *RmLocalFileFix) GetRuleId() string {
	return r.RuleId
}

func (r *RmLocalFileFix) Type() string {
	return "rm_local_file"
}

func (r *RmLocalFileFix) ApplyFix() error {
	fs := FsFactory()
	var err error
	for _, path := range r.Paths {
		removeErr := fs.Remove(path)
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
