package pkg

import (
	"fmt"
	"io/fs"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/afero"
	"github.com/zclconf/go-cty/cty"
)

var _ Fix = &LocalFileFix{}

type LocalFileFix struct {
	*BaseBlock
	*BaseFix
	Paths   []string `json:"paths" hcl:"paths"`
	Content string   `json:"content" hcl:"content"`
	Mode    *uint32  `json:"mode" hcl:"mode,optional"`
}

func (lf *LocalFileFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"paths":   ToCtyValue(lf.Paths),
		"content": ToCtyValue(lf.Content),
		"mode":    ToCtyValue(lf.Mode),
	}
}

func (lf *LocalFileFix) Type() string {
	return "local_file"
}

func (lf *LocalFileFix) Apply() error {
	var err error
	var filemode = fs.FileMode(0644)
	if lf.Mode != nil {
		mode, err := strconv.ParseUint(strconv.Itoa(int(*lf.Mode)), 8, 32)
		if err != nil {
			return fmt.Errorf("invalid file mode: %w", err)
		}
		filemode = fs.FileMode(mode)
	}

	for _, path := range lf.Paths {
		writeErr := afero.WriteFile(FsFactory(), path, []byte(lf.Content), filemode)
		if writeErr != nil {
			err = multierror.Append(err, writeErr)
		}
	}

	return err
}
