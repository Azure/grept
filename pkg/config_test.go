package pkg

import (
	"fmt"
	"github.com/spf13/afero"
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
		rule_id = rule.file_hash.sample.id
		path = "/path/to/file.txt"  
		content = "Hello, world!"
	}  
	`

	config, err := ParseConfig("", "config.hcl", content)
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
	assert.Regexp(t, `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`, lff.RuleId)
	assert.Equal(t, "/path/to/file.txt", lff.Path)
	assert.Equal(t, "Hello, world!", lff.Content)
}

func TestUnregisteredFix(t *testing.T) {
	hcl := `  
	fix "unregistered_fix" sample {  
		rule_id = "c01d7cf6-ec3f-47f0-9556-a5d6e9009a43"  
		path = "/path/to/file.txt"  
		content = "Hello, world!"  
	}  
	`

	_, err := ParseConfig("", "test.hcl", hcl)
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

	_, err := ParseConfig("", "test.hcl", hcl)
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

	_, err := ParseConfig("", "test.hcl", hcl)
	require.NotNil(t, err)

	expectedError := "invalid block type: invalid_block"
	assert.Contains(t, err.Error(), expectedError)
}

func TestEvalContextRef(t *testing.T) {
	hcl := `
	rule "file_hash" sample {  
		glob = "LICENSE"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_id = rule.file_hash.sample.id
		path = rule.file_hash.sample.glob  
		content = "Hello, world!"
	}
`
	config, err := ParseConfig("", "config.hcl", hcl)
	assert.NoError(t, err)
	require.Equal(t, 1, len(config.Fixes))
	fix := config.Fixes[0].(*LocalFile)
	assert.Equal(t, "LICENSE", fix.Path)
}

func TestFunctionInEvalContext(t *testing.T) {
	// Create a in-memory filesystem
	fs := afero.NewMemMapFs()

	// Create a file with some content
	fileContent := "Hello, world!"
	err := afero.WriteFile(fs, "/testfile", []byte(fileContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Define a configuration string with a rule block that uses the md5 function
	configStr := fmt.Sprintf(`  
	rule "file_hash" "test_rule" {  
		glob = "/testfile"  
		hash = md5("%s")  
		algorithm = "md5"  
	}  
	`, fileContent)

	// Parse the configuration string
	config, err := ParseConfig(".", "test.hcl", configStr)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 1, len(config.Rules))
	rule, ok := config.Rules[0].(*FileHashRule)
	require.True(t, ok)
	rule.fs = fs
	err = rule.Check()
	assert.NoError(t, err)
}
