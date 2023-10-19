package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/hashicorp/go-getter/v2"
	"github.com/spf13/afero"
	"os"
)

func getConfigFolder(path string, ctx context.Context) (configPath string, onDefer func(), err error) {
	fs := pkg.FsFactory()
	exists, err := afero.Exists(fs, path)
	if exists && err == nil {
		return path, nil, nil
	}
	tmp, err := os.MkdirTemp("", "grept")
	if err != nil {
		return "", nil, err
	}
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
