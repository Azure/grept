package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/google/uuid"
	"github.com/hashicorp/go-getter/v2"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

func getConfigFolder(path string, ctx context.Context) (configPath string, onDefer func(), err error) {
	exists, err := isReadableLocalPath(path)
	if exists && err == nil {
		return path, nil, nil
	}
	absPath, err := filepath.Abs(path)
	if err == nil {
		exists, err = isReadableLocalPath(absPath)
		if exists && err == nil {
			return path, nil, nil
		}
	}
	tmp := filepath.Join(os.TempDir(), uuid.NewString())
	cleaner := func() {
		_ = os.RemoveAll(tmp)
	}
	result, err := getter.Get(ctx, tmp, path)
	if err != nil {
		return "", cleaner, err
	}
	if result == nil {
		return "", cleaner, fmt.Errorf("cannot get config path")
	}
	return result.Dst, cleaner, nil
}

func isReadableLocalPath(path string) (bool, error) {
	fs := pkg.FsFactory()
	return afero.Exists(fs, path)
}
