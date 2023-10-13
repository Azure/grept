package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

type LocalFile struct {
	*BaseFix
	Path    string
	Content string
}

func (lf *LocalFile) Value() cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"path":    cty.StringVal(lf.Path),
		"content": cty.StringVal(lf.Content),
	})
}

func (lf *LocalFile) Type() string {
	return "local_file"
}

var _ Fix = &LocalFile{}

func (lf *LocalFile) ApplyFix() error {
	err := afero.WriteFile(fsFactory(), lf.Path, []byte(lf.Content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (lf *LocalFile) Parse(b *hclsyntax.Block) error {
	err := lf.BaseFix.Parse(b)
	if err != nil {
		return err
	}
	lf.Path, err = readRequiredStringAttribute(b, "path", lf.EvalContext())
	if err != nil {
		return err
	}
	lf.Content, err = readRequiredStringAttribute(b, "content", lf.EvalContext())
	return err
}
