package pkg

import (
	"context"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type dirExistRuleSuite struct {
	suite.Suite
	*testBase
}

func TestDirExistRuleSuite(t *testing.T) {
	suite.Run(t, new(dirExistRuleSuite))
}

func (s *dirExistRuleSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *dirExistRuleSuite) TearDownTest() {
	s.teardown()
}

func (s *dirExistRuleSuite) TestConfig_DirExistRule() {
	fs := s.fs
	t := s.T()
	hcl := `  
rule dir_exist test {  
  dir = "./testdir"  
}`

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

func (s *dirExistRuleSuite) TestConfig_DirExistRule_CheckFailed() {
	hcl := `  
rule dir_exist test {  
  dir = "./nonexistent"  
}`

	fs := s.fs
	t := s.T()
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
