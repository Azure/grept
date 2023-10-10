package fixes

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLocalFile_ApplyFix_CreateNewFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	fix := &LocalFile{
		fs:      fs,
		Rule:    "rule1",
		Path:    "/file1.txt",
		Content: "Hello, world!",
	}

	err := fix.ApplyFix()
	assert.NoError(t, err)

	// Check that the file was created with the correct content
	content, err := afero.ReadFile(fix.fs, fix.Path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

func TestLocalFile_ApplyFix_OverwriteExistingFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	fix := &LocalFile{
		fs:      fs,
		Rule:    "rule1",
		Path:    "/file1.txt",
		Content: "Hello, world!",
	}

	// Create the file first
	err := fix.ApplyFix()
	assert.NoError(t, err)

	// Now overwrite it
	fix.Content = "New content"
	err = fix.ApplyFix()
	assert.NoError(t, err)

	// Check that the file was overwritten with the correct content
	content, err := afero.ReadFile(fix.fs, fix.Path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}
