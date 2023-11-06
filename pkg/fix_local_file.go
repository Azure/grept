package pkg

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &LocalFileFix{}

type LocalFileFix struct {
	*BaseBlock
	baseFix
	RuleId  string   `json:"rule_id" hcl:"rule_id"`
	Paths   []string `json:"paths" hcl:"paths"`
	Content string   `json:"content" hcl:"content"`
}

func (lf *LocalFileFix) GetRuleId() string {
	return lf.RuleId
}

func (lf *LocalFileFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_id": ToCtyValue(lf.RuleId),
		"paths":   ToCtyValue(lf.Paths),
		"content": ToCtyValue(lf.Content),
	}
}

func (lf *LocalFileFix) Type() string {
	return "local_file"
}

func (lf *LocalFileFix) ApplyFix() error {
	var err error
	for _, path := range lf.Paths {
		writeErr := afero.WriteFile(FsFactory(), path, []byte(lf.Content), 0644)
		if writeErr != nil {
			err = multierror.Append(err, writeErr)
		}
	}

	return err
}
