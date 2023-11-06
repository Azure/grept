package pkg

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"path/filepath"
	"testing"
)

type fixRmLocalFileSuite struct {
	suite.Suite
	*testBase
}

func TestFixRmLocalFileSuite(t *testing.T) {
	suite.Run(t, new(fixRmLocalFileSuite))
}

func (s *fixRmLocalFileSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *fixRmLocalFileSuite) TearDownTest() {
	s.teardown()
}

func (s *fixRmLocalFileSuite) TestRemoveFile_FileExist() {
	fs := s.fs
	t := s.T()
	tmpFile, err := afero.TempFile(fs, "", "test")
	require.NoError(t, err)
	fileName := tmpFile.Name()
	rf := &RmLocalFileFix{
		BaseBlock: &BaseBlock{
			c: &Config{},
		},
		Paths: []string{fileName},
	}

	err = rf.ApplyFix()

	require.NoError(t, err)
	exists, err := afero.Exists(fs, fileName)
	require.NoError(t, err)
	assert.False(t, exists)
}

func (s *fixRmLocalFileSuite) TestRemoveFile_FileNotExist() {
	t := s.T()
	fileName := "/path/to/not-exist-file"
	rf := &RmLocalFileFix{
		BaseBlock: &BaseBlock{
			c: &Config{},
		},
		Paths: []string{fileName},
	}

	err := rf.ApplyFix()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), filepath.FromSlash(fileName))
}
