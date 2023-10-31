package pkg

import (
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenameFile_ApplyFix(t *testing.T) {
	fs := afero.NewMemMapFs()

	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()

	// Create temporary file for testing
	tmpFile, err := afero.TempFile(fs, "", "test")
	assert.NoError(t, err)
	oldName := tmpFile.Name()
	newName := oldName + "_renamed"

	// Initialize RenameFile fix
	rf := &RenameFile{
		BaseFix: &BaseFix{
			c: &Config{},
		},
		OldName: oldName,
		NewName: newName,
	}

	// Apply the fix
	err = rf.ApplyFix()

	// Assert there is no error and the renamed file exists
	assert.NoError(t, err)
	exists, err := afero.Exists(fs, newName)
	require.NoError(t, err)
	assert.True(t, exists)
	exists, err = afero.Exists(fs, oldName)
	require.NoError(t, err)
	assert.False(t, exists)
}
