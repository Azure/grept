package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	content := `  
	rule "file_hash" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule = file_hash.sample
		path = "/path/to/file.txt"  
		content = "Hello, world!"
	}  
	`

	config, err := ParseConfig("config.hcl", content)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Rules))
	fhr, ok := config.Rules[0].(*FileHashRule)
	require.True(t, ok)
	assert.Equal(t, "*.txt", fhr.Glob)
	assert.Equal(t, "abc123", fhr.Hash)
	assert.Equal(t, "sha256", fhr.Algorithm)

	assert.Equal(t, 1, len(config.Fixes))
	lff, ok := config.Fixes[0].(*LocalFile)
	require.True(t, ok)
	assert.Equal(t, "file_hash.sample", lff.Rule)
	assert.Equal(t, "/path/to/file.txt", lff.Path)
	assert.Equal(t, "Hello, world!", lff.Content)
}

func TestUnregisteredFix(t *testing.T) {
	hcl := `  
	fix "unregistered_fix" sample {  
		rule = "file_hash.sample"  
		path = "/path/to/file.txt"  
		content = "Hello, world!"  
	}  
	`

	_, err := ParseConfig("test.hcl", hcl)
	require.NotNil(t, err)
	expectedError := "unregistered fix: unregistered_fix"
	assert.Contains(t, err.Error(), expectedError)
}

func TestUnregisteredRule(t *testing.T) {
	hcl := `  
	rule "unregistered_rule" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
	`

	_, err := ParseConfig("test.hcl", hcl)
	require.NotNil(t, err)

	expectedError := "unregistered rule: unregistered_rule"
	assert.Contains(t, err.Error(), expectedError)
}

func TestInvalidBlockType(t *testing.T) {
	hcl := `  
	invalid_block "invalid_type" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
	`

	_, err := ParseConfig("test.hcl", hcl)
	require.NotNil(t, err)

	expectedError := "invalid block type: invalid_block"
	assert.Contains(t, err.Error(), expectedError)
}
