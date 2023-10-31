package pkg

import (
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGitIgnore_Load(t *testing.T) {
	// Set up a in-memory filesystem with a .gitignore file
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()

	ignoreContent := "# This is a comment\n*.log\n"
	_ = afero.WriteFile(fs, ".gitignore", []byte(ignoreContent), 0644)

	// create GitIgnore instance and load .gitignore content
	gitIgnore := &GitIgnore{
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

func TestGitIgnore_NoGitIgnoreFile(t *testing.T) {
	// Set up a in-memory filesystem with a .gitignore file
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()

	// create GitIgnore instance and load .gitignore content
	gitIgnore := &GitIgnore{
		BaseData: &BaseData{
			c: &Config{},
		},
	}

	err := gitIgnore.Load()
	require.NoError(t, err)
	assert.Len(t, gitIgnore.Records, 0)
}
