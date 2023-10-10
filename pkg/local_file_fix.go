package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
)

type LocalFile struct {
	*BaseFix
	fs      afero.Fs
	Path    string `hcl:"path,attr"`
	Content string `hcl:"content,attr"`
}

var _ Fix = &LocalFile{}

func (lf *LocalFile) ApplyFix() error {
	err := afero.WriteFile(lf.fs, lf.Path, []byte(lf.Content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (lf *LocalFile) Validate() error {
	// Check conditions specific to LocalFile
	if lf.Path == "" {
		return fmt.Errorf("path is required")
	}
	// You can also call the Validate method of BaseFix if needed
	return lf.BaseFix.Validate()
}

func (lf *LocalFile) Parse(b *hclsyntax.Block) error {
	err := lf.BaseFix.Parse(b)
	if err != nil {
		return err
	}
	lf.Path, err = readRequiredStringAttribute(b, "path", lf.ctx)
	if err != nil {
		return err
	}
	lf.Content, err = readRequiredStringAttribute(b, "content", lf.ctx)
	return err
}
