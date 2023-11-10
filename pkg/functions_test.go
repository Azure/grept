package pkg

import (
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFunction_Getenv(t *testing.T) {
	// Create a mock file system and write a file
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs { return fs })
	defer stub.Reset()
	content := "Expected content"
	t.Setenv("TEST_URL", content)

	// Define a sample config for testing
	sampleConfig := `  
	data "http" "foo" {  
		url = env("TEST_URL")
	}  
	`
	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(sampleConfig), 0644)
	// Parse the config
	config, err := ParseConfig(".", nil)
	require.NoError(t, err)
	http := config.DataSources[0].(*HttpDatasource)
	assert.Equal(t, content, http.Url)
}

func TestFunction_Compliment(t *testing.T) {
	// Create a mock file system and write a file
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs { return fs })
	defer stub.Reset()
	// Define a sample config for testing
	sampleConfig := `  
	rule "must_be_true" "one" {  
		condition = contains(compliment([2, 4, 6, 8, 10, 12], [4, 6, 8], [12]), 2)
	}
	rule "must_be_true" "two" {  
		condition = contains(compliment([2, 4, 6, 8, 10, 12], [4, 6, 8], [12]), 10)
	}
	rule "must_be_true" "three" {  
		condition = length(compliment([2, 4, 6, 8, 10, 12], [4, 6, 8], [12])) == 2
	}
	`
	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(sampleConfig), 0644)
	// Parse the config
	config, err := ParseConfig(".", nil)
	require.NoError(t, err)
	for _, rule := range config.Rules {
		err, rErr := rule.Check()
		assert.NoError(t, rErr)
		assert.NoError(t, err)
	}
}
