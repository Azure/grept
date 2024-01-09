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
	Paths   []string     `json:"paths" hcl:"paths"`
	Content string       `json:"content" hcl:"content"`
	Mode    *fs.FileMode `json:"mode" hcl:"mode,optional" default:"0644" validate:"file_mode"`
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
	fm, err := toDecimal(*lf.Mode)
	if err != nil {
		return err
	}

	for _, path := range lf.Paths {
		writeErr := afero.WriteFile(FsFactory(), path, []byte(lf.Content), fm)
		if writeErr != nil {
			err = multierror.Append(err, writeErr)
		}
	}

	return err
}

func toDecimal(octalMode fs.FileMode) (fs.FileMode, error) {
	mode, err := strconv.ParseUint(strconv.Itoa(int(octalMode)), 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode: %w", err)
	}
	fm := fs.FileMode(mode)
	return fm, nil
}
