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
	RuleIds []string `json:"rule_ids" hcl:"rule_ids"`
	Paths   []string `json:"paths" hcl:"paths"`
	Content string   `json:"content" hcl:"content"`
}

func (lf *LocalFileFix) GetRuleIds() []string {
	return lf.RuleIds
}

func (lf *LocalFileFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_ids": ToCtyValue(lf.RuleIds),
		"paths":    ToCtyValue(lf.Paths),
		"content":  ToCtyValue(lf.Content),
	}
}

func (lf *LocalFileFix) Type() string {
	return "local_file"
}

func (lf *LocalFileFix) Execute() error {
	var err error
	for _, path := range lf.Paths {
		writeErr := afero.WriteFile(FsFactory(), path, []byte(lf.Content), 0644)
		if writeErr != nil {
			err = multierror.Append(err, writeErr)
		}
	}

	return err
}
