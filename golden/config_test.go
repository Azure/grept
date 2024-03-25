package golden

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ Config = &DummyConfig{}

type DummyConfig struct {
	*BaseConfig
}

func NewDummyConfig(baseDir string, ctx context.Context, hclBlocks []*HclBlock) (Config, error) {
	cfg := &DummyConfig{
		BaseConfig: NewBasicConfig(baseDir, ctx),
	}
	return cfg, InitConfig(cfg, hclBlocks)
}

func BuildDummyConfig(baseDir, cfgDir string, ctx context.Context) (Config, error) {
	var err error
	hclBlocks, err := loadHclBlocks(false, cfgDir)
	if err != nil {
		return nil, err
	}

	c, err := NewDummyConfig(baseDir, ctx, hclBlocks)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func loadHclBlocks(ignoreUnsupportedBlock bool, dir string) ([]*HclBlock, error) {
	fs := testFsFactory()
	matches, err := afero.Glob(fs, filepath.Join(dir, "*.hcl"))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no `.hcl` file found at %s", dir)
	}

	var blocks []*HclBlock

	for _, filename := range matches {
		content, fsErr := afero.ReadFile(fs, filename)
		if fsErr != nil {
			err = multierror.Append(err, fsErr)
			continue
		}
		file, diag := hclsyntax.ParseConfig(content, filename, hcl.InitialPos)
		if diag.HasErrors() {
			err = multierror.Append(err, diag.Errs()...)
			continue
		}
		body := file.Body.(*hclsyntax.Body)
		blocks = append(blocks, AsHclBlocks(body.Blocks)...)
	}
	if err != nil {
		return nil, err
	}

	var r []*HclBlock

	// First loop: parse all rule blocks
	for _, b := range blocks {
		if IsBlockTypeWanted(b.Type) {
			r = append(r, b)
			continue
		}
		if !ignoreUnsupportedBlock {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", b.Type, b.Range().String()))
		}
	}
	return r, err
}

func RunDummyPlan(c Config) (*DummyPlan, error) {
	err := c.RunPlan()
	if err != nil {
		return nil, err
	}

	return &DummyPlan{
		Datas:     Blocks[TestData](c),
		Resources: Blocks[TestResource](c),
	}, nil
}

type DummyPlan struct {
	Datas     []TestData
	Resources []TestResource
}

type configSuite struct {
	suite.Suite
	*testBase
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(configSuite))
}

func (s *configSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *configSuite) TearDownTest() {
	s.teardown()
}

func (s *configSuite) TestParseConfig() {
	content := `
	data "dummy" sample {
		data = {
          key = "value"
        }
	}

	resource "dummy" hello_world {
		tags = data.dummy.sample.data
	}
	`

	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	t := s.T()

	config, err := BuildDummyConfig("", "", nil)
	require.NoError(t, err)
	_, err = RunDummyPlan(config)
	require.NoError(t, err)
	dataBlocks := Blocks[TestData](config)
	assert.Len(t, dataBlocks, 1)
	dummyData, ok := dataBlocks[0].(*DummyData)
	require.True(t, ok)
	assert.Equal(t, map[string]string{
		"key": "value",
	}, dummyData.Tags)

	resources := Blocks[TestResource](config)
	assert.Len(t, resources, 1)
	res, ok := resources[0].(*DummyResource)
	require.True(t, ok)
	assert.Equal(t, map[string]string{
		"key": "value",
	}, res.Tags)
}

func (s *configSuite) TestUnregisteredBlock() {
	hcl := `
	data "unregistered_data" sample {
		path = "/path/to/file.txt"
	}
	`

	t := s.T()
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{hcl})
	_, err := BuildDummyConfig("", "", nil)
	require.NotNil(t, err)
	expectedError := "unregistered data: unregistered_data"
	assert.Contains(t, err.Error(), expectedError)
}

func (s *configSuite) TestInvalidBlockType() {
	hcl := `
	invalid_block "invalid_type" sample {
		glob = "*.txt"
		hash = "abc123"
		algorithm = "sha256"
	}
	`

	t := s.T()
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{hcl})
	_, err := BuildDummyConfig("", "", nil)
	require.NotNil(t, err)

	expectedError := "invalid block type: invalid_block"
	assert.Contains(t, err.Error(), expectedError)
}

func (s *configSuite) TestFunctionInEvalContext() {
	t := s.T()
	configStr := `
	data "dummy" "foo" {
		data = {
			key = trim("?!hello?!", "!?")
		}
	}
	`
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{configStr})

	config, err := BuildDummyConfig("/", ".", nil)
	require.NoError(t, err)
	_, err = RunDummyPlan(config)
	require.NoError(t, err)
	ds := Blocks[TestData](config)
	require.Len(t, ds, 1)
	data, ok := ds[0].(*DummyData)
	require.True(t, ok)
	assert.Equal(t, "hello", data.Tags["key"])
}

func (s *configSuite) TestLocalsBlockShouldBeParsedIntoMultipleLocalBlocks() {
	code := `
locals {
  a = "a"
  b = 1
}
`
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{code})
	c, err := BuildDummyConfig("/", "", nil)
	s.NoError(err)
	locals := Blocks[Local](c)
	s.Len(locals, 2)
}

func (s *configSuite) TestForEach_ForEachBlockShouldBeExpanded() {
	hclConfig := `
	locals {
		items = ["item1", "item2", "item3"]
	}

	data "dummy" "foo" {
		for_each = local.items
	}
`
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{hclConfig})

	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	s.Len(Blocks[TestData](config), 3)
}

func (s *configSuite) TestForEachAndAddressIndex() {
	hclConfig := `
    locals {
        items = toset(["item1", "item2", "item3"])
    }

    data "dummy" foo {
        for_each = local.items
        data = {
			key = each.value
		}
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{hclConfig})

	config, err := BuildDummyConfig("", "", nil)
	require.NoError(s.T(), err)

	p, err := RunDummyPlan(config)
	require.NoError(s.T(), err)
	s.Len(p.Datas, 3)
	values := make(map[string]string)
	for _, data := range p.Datas {
		values[data.Address()] = data.(*DummyData).Tags["key"]
	}
	s.Equal(map[string]string{
		`data.dummy.foo[item1]`: "item1",
		`data.dummy.foo[item2]`: "item2",
		`data.dummy.foo[item3]`: "item3",
	}, values)
}

func (s *configSuite) TestForEach_forEachAsToggle() {
	hclConfig := `
    locals {
        items = toset(["item1", "item2", "item3"])
    }

    data "dummy" sample {
        for_each = false ? locals.items : []
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{hclConfig})

	config, err := BuildDummyConfig("", "", nil)
	require.NoError(s.T(), err)
	s.Len(Blocks[TestData](config), 0)
}

func (s *configSuite) TestForEach_blocksWithIndexShouldHasNewBlockId() {
	hclConfig := `
    locals {
        items = toset(["item1", "item2"])
    }

    data "dummy" foo {
        for_each = local.items
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{hclConfig})

	config, err := BuildDummyConfig("", "", nil)
	require.NoError(s.T(), err)
	ds := Blocks[TestData](config)
	s.Len(ds, 2)
	ruleBlocks := ds
	d0 := ruleBlocks[0].(Block)
	d1 := ruleBlocks[1].(Block)
	s.NotEqual(d0.Id(), d1.Id())
}
