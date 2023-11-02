package pkg

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type gitIgnoreSuite struct {
	suite.Suite
	*testBase
}

func (s *gitIgnoreSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *gitIgnoreSuite) TearDownTest() {
	s.teardown()
}

func TestGitIgnoreData(t *testing.T) {
	suite.Run(t, new(gitIgnoreSuite))
}

func (s *gitIgnoreSuite) TestGitIgnore_Load() {
	// Set up a in-memory filesystem with a .gitignore file
	fs := s.fs
	t := s.T()
	ignoreContent := "# This is a comment\n*.log\n"
	_ = afero.WriteFile(fs, ".gitignore", []byte(ignoreContent), 0644)

	// create GitIgnoreDatasource instance and load .gitignore content
	gitIgnore := &GitIgnoreDatasource{
		BaseData: &BaseData{
			c: &Config{},
		},
	}

	err := gitIgnore.Load()

	require.NoError(t, err)

	// only non-comment lines should be in Records
	assert.Len(t, gitIgnore.Records, 1)
	assert.Equal(t, "*.log", gitIgnore.Records[0])
}

func (s *gitIgnoreSuite) TestGitIgnore_NoGitIgnoreFile() {
	t := s.T()

	// create GitIgnoreDatasource instance and load .gitignore content
	gitIgnore := &GitIgnoreDatasource{
		BaseData: &BaseData{
			c: &Config{},
		},
	}

	err := gitIgnore.Load()
	require.NoError(t, err)
	assert.Len(t, gitIgnore.Records, 0)
}

func (s *gitIgnoreSuite) TestGitIgnore_TabSpaceNewLine() {
	fs := s.fs
	t := s.T()

	content := "\t\n   \n \t \n\t \t\n\n\r\n"
	_ = afero.WriteFile(fs, ".gitignore", []byte(content), 0644)

	gitIgnore := &GitIgnoreDatasource{
		BaseData: &BaseData{
			c: &Config{},
		},
	}

	err := gitIgnore.Load()

	require.NoError(t, err)
	assert.Empty(t, gitIgnore.Records)
}
