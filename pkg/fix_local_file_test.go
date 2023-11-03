package pkg

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type localFileFixSuite struct {
	suite.Suite
	*testBase
}

func TestLocalFileFix(t *testing.T) {
	suite.Run(t, new(localFileFixSuite))
}

func (s *localFileFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *localFileFixSuite) TearDownTest() {
	s.teardown()
}

func (s *localFileFixSuite) TestLocalFile_ApplyFix_CreateNewFile() {
	fs := s.fs
	t := s.T()
	fix := &LocalFileFix{
		Paths:   []string{"/file1.txt"},
		Content: "Hello, world!",
	}

	err := fix.ApplyFix()
	assert.NoError(t, err)

	// Check that the file was created with the correct content
	content, err := afero.ReadFile(fs, fix.Paths[0])
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

func (s *localFileFixSuite) TestLocalFile_ApplyFix_OverwriteExistingFile() {
	fs := s.fs
	t := s.T()
	path := "/file1.txt"
	fix := &LocalFileFix{
		baseBlock: &baseBlock{},
		RuleId:    uuid.NewString(),
		Paths:     []string{path},
		Content:   "Hello, world!",
	}

	// Create the file first
	err := fix.ApplyFix()
	assert.NoError(t, err)

	// Now overwrite it
	fix.Content = "New content"
	err = fix.ApplyFix()
	assert.NoError(t, err)

	// Check that the file was overwritten with the correct content
	content, err := afero.ReadFile(fs, path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

func (s *localFileFixSuite) TestLocalFile_ApplyFix_FileInSubFolder() {
	fs := s.fs
	t := s.T()
	path := "/example/sub1/file1.txt"
	fix := &LocalFileFix{
		baseBlock: &baseBlock{},
		RuleId:    uuid.NewString(),
		Paths:     []string{path},
		Content:   "Hello, world!",
	}

	// Create the file first
	err := fix.ApplyFix()
	assert.NoError(t, err)

	// Now overwrite it
	fix.Content = "New content"
	err = fix.ApplyFix()
	assert.NoError(t, err)

	// Check that the file was overwritten with the correct content
	content, err := afero.ReadFile(fs, path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}
