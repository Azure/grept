package pkg

import (
	"context"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_DirExistRule(t *testing.T) {
	hcl := `  
rule dir_exist test {  
  dir = "./testdir"  
}`

	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	err := fs.Mkdir("./testdir", 0755)
	require.NoError(t, err)

	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(hcl), 0644)
	c, err := ParseConfig(".", context.Background())
	require.NoError(t, err)

	assert.Len(t, c.Rules, 1)

	rule, ok := c.Rules[0].(*DirExistRule)
	assert.True(t, ok)
	assert.Equal(t, "./testdir", rule.Dir)
	checkError, runtimeError := rule.Check()
	assert.NoError(t, checkError)
	assert.NoError(t, runtimeError)
}

func TestConfig_DirExistRule_CheckFailed(t *testing.T) {
	hcl := `  
rule dir_exist test {  
  dir = "./nonexistent"  
}`

	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(hcl), 0644)
	c, err := ParseConfig(".", context.Background())
	if err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	assert.Len(t, c.Rules, 1)

	rule, ok := c.Rules[0].(*DirExistRule)
	assert.True(t, ok)
	assert.Equal(t, "./nonexistent", rule.Dir)

	checkError, runtimeError := rule.Check()
	assert.Error(t, checkError)
	assert.NoError(t, runtimeError)
}
