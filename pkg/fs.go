package pkg

import "github.com/spf13/afero"

var FsFactory = func() afero.Fs {
	return afero.NewOsFs()
}
