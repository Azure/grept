package pkg

import (
	"context"
	"fmt"
	"github.com/Azure/golden"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"path/filepath"
)

var _ golden.Config = &GreptConfig{}

type GreptConfig struct {
	*golden.BaseConfig
}

func NewGreptConfig(baseDir string, cliFlagAssignedVariables []golden.CliFlagAssignedVariables, ctx context.Context, hclBlocks []*golden.HclBlock) (golden.Config, error) {
	cfg := &GreptConfig{
		BaseConfig: golden.NewBasicConfig(baseDir, "grept", "grept", nil, cliFlagAssignedVariables, ctx),
	}
	return cfg, golden.InitConfig(cfg, hclBlocks)
}

func BuildGreptConfig(baseDir, cfgDir string, ctx context.Context, cliFlagAssignedVariables []golden.CliFlagAssignedVariables) (golden.Config, error) {
	var err error
	hclBlocks, err := loadGreptHclBlocks(false, cfgDir)
	if err != nil {
		return nil, err
	}

	c, err := NewGreptConfig(baseDir, cliFlagAssignedVariables, ctx, hclBlocks)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func loadGreptHclBlocks(ignoreUnsupportedBlock bool, dir string) ([]*golden.HclBlock, error) {
	fs := FsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.grept.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.grept.hcl` file found at %s", dir)
	}

	var blocks []*golden.HclBlock

	for _, filename := range matches {
		content, fsErr := afero.ReadFile(fs, filename)
		if fsErr != nil {
			err = multierror.Append(err, fsErr)
			continue
		}
		readFile, diag := hclsyntax.ParseConfig(content, filename, hcl.InitialPos)
		if diag.HasErrors() {
			err = multierror.Append(err, diag.Errs()...)
			continue
		}
		writeFile, _ := hclwrite.ParseConfig(content, filename, hcl.InitialPos)
		readBody := readFile.Body.(*hclsyntax.Body)
		writeBody := writeFile.Body()
		blocks = append(blocks, golden.AsHclBlocks(readBody.Blocks, writeBody.Blocks())...)
	}
	if err != nil {
		return nil, err
	}

	var r []*golden.HclBlock

	// First loop: parse all rule blocks
	for _, b := range blocks {
		if golden.IsBlockTypeWanted(b.Type) {
			r = append(r, b)
			continue
		}
		if !ignoreUnsupportedBlock {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", b.Type, b.Range().String()))
		}
	}
	return r, err
}
