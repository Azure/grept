package pkg

import (
	"github.com/google/uuid"
	"github.com/prashantv/gostub"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestLocalFile_ApplyFix_CreateNewFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	fix := &LocalFile{
		Path:    "/file1.txt",
		Content: "Hello, world!",
	}

	err := fix.ApplyFix()
	assert.NoError(t, err)

	// Check that the file was created with the correct content
	content, err := afero.ReadFile(fs, fix.Path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}

func TestLocalFile_ApplyFix_OverwriteExistingFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	fix := &LocalFile{
		BaseFix: &BaseFix{RuleId: uuid.NewString()},
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
	content, err := afero.ReadFile(fs, fix.Path)
	assert.NoError(t, err)
	assert.Equal(t, fix.Content, string(content))
}
