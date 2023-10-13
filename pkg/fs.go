package pkg

import "github.com/spf13/afero"

var fsFactory = func() afero.Fs {
	return afero.NewOsFs()
}
