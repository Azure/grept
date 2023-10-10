package pkg

import (
	"testing"

	"github.com/Azure/grept/pkg/fixes"
	"github.com/Azure/grept/pkg/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	content := `  
	rule "file_hash" {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" {  
		rule = "file_hash_rule"  
		path = "/path/to/file.txt"  
		content = "Hello, world!"
	}  
	`

	config, err := ParseConfig("config.hcl", content)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Rules))
	fhr, ok := config.Rules[0].(*rules.FileHashRule)
	require.True(t, ok)
	assert.Equal(t, "*.txt", fhr.Glob)
	assert.Equal(t, "abc123", fhr.Hash)
	assert.Equal(t, "sha256", fhr.Algorithm)

	assert.Equal(t, 1, len(config.Fixes))
	lff, ok := config.Fixes[0].(*fixes.LocalFile)
	require.True(t, ok)
	assert.Equal(t, "file_hash_rule", lff.Rule)
	assert.Equal(t, "/path/to/file.txt", lff.Path)
	assert.Equal(t, "Hello, world!", lff.Content)
}
