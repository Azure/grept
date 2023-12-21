package pkg

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &CopyFileFix{}

type CopyFileFix struct {
	*BaseBlock
	*BaseFix
	Src  string `json:"src" hcl:"src"`
	Dest string `json:"dest" hcl:"dest"`
}

func (c *CopyFileFix) Type() string {
	return "copy_file"
}

func (c *CopyFileFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"rule_ids": ToCtyValue(c.RuleIds),
		"src":      ToCtyValue(c.Src),
		"dest":     ToCtyValue(c.Dest),
	}
}

func (c *CopyFileFix) Apply() error {
	fs := FsFactory()
	file, err := fs.Open(c.Src)
	if err != nil {
		return fmt.Errorf("error on reading src %s %+v", c.Src, err)
	}
	defer func() {
		_ = file.Close()
	}()
	err = afero.WriteReader(fs, c.Dest, file)
	if err != nil {
		return fmt.Errorf("error on writing dest %s %+v", c.Dest, err)
	}
	return nil
}
