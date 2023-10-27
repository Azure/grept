package pkg

import (
	"context"
	"fmt"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"net/http"
	"net/http/httptest"
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
		paths = ["/path/to/file.txt"]  
		content = "Hello, world!"
	}  
	`

	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	defer stub.Reset()

	config, err := ParseConfig("", nil)
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
	assert.Equal(t, "/path/to/file.txt", lff.Paths[0])
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

	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	defer stub.Reset()
	_, err := ParseConfig(".", nil)
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

	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	defer stub.Reset()
	_, err := ParseConfig(".", nil)
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

	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	defer stub.Reset()
	_, err := ParseConfig("", nil)
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
		paths = [rule.file_hash.sample.glob]  
		content = "Hello, world!"
	}
`
	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	defer stub.Reset()
	config, err := ParseConfig("", nil)
	assert.NoError(t, err)
	require.Equal(t, 1, len(config.Fixes))
	fix := config.Fixes[0].(*LocalFile)
	assert.Equal(t, "LICENSE", fix.Paths[0])
}

func TestFunctionInEvalContext(t *testing.T) {
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs {
		return fs
	})
	defer stub.Reset()
	fileContent := "Hello, world!"
	_ = afero.WriteFile(fs, "/testfile", []byte(fileContent), 0644)

	configStr := fmt.Sprintf(`  
	rule "file_hash" "test_rule" {  
		glob = "/testfile"  
		hash = md5("%s")  
		algorithm = "md5"  
	}  
	`, fileContent)
	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(configStr), 0644)

	config, err := ParseConfig(".", nil)
	if err != nil {
		t.Fatal(err)
	}

	require.Equal(t, 1, len(config.Rules))
	rule, ok := config.Rules[0].(*FileHashRule)
	require.True(t, ok)
	_, err = rule.Check()
	assert.NoError(t, err)
}

func TestParseConfigHttpBlock(t *testing.T) {
	// Define a HCL configuration with an http block
	hclConfig := `  
	data "http" "example" {  
		url = "http://example.com"  
		method = "GET"  
		request_body = "Hello"  
		request_headers = {  
			"Content-Type" = "application/json"  
			"Accept" = "application/json"  
		}  
	}  
	`

	dir := "."
	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})
	defer stub.Reset()

	// Parse the configuration
	config, err := ParseConfig(dir, nil)
	assert.NoError(t, err, "ParseConfig should not return an error")

	// Check the parsed configuration
	assert.Equal(t, 1, len(config.DataSources), "There should be one data source")

	httpData, ok := config.DataSources[0].(*HttpDatasource)
	assert.True(t, ok)
	assert.Equal(t, "http://example.com", httpData.Url)
	assert.Equal(t, "GET", httpData.Method)
	assert.Equal(t, "Hello", httpData.RequestBody)
	assert.Equal(t, map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}, httpData.RequestHeaders)
	assert.Equal(t, "example", httpData.Name())
}

func TestPlanError_DatasourceError(t *testing.T) {
	// Create a mock HTTP server that always returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Mock server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`  
	data "http" "foo" {  
		url = "%s"  
	}  
`, server.URL)
	stub := dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})
	defer stub.Reset()
	// Parse the config
	config, err := ParseConfig(".", nil)
	require.NoError(t, err)

	config.ctx = context.TODO()

	// Test the Plan method
	plan, err := config.Plan()
	assert.Empty(t, plan)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error making request")
	assert.Contains(t, err.Error(), "data.http.foo")
}

func TestPlanError_FileHashRuleError(t *testing.T) {
	// Create a mock HTTP server that returns a specific content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	// Create a mock file system and write a file
	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs { return fs })
	defer stub.Reset()

	_ = afero.WriteFile(fs, "/testfile", []byte("Different content"), 0644)

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`  
	data "http" "foo" {  
		url = "%s"  
	}  
  
	rule "file_hash" "bar" {  
		glob = "/testfile"  
		hash = md5(data.http.foo.response_body)  
		algorithm = "md5"  
	}  
	`, server.URL)
	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(sampleConfig), 0644)
	// Parse the config
	config, err := ParseConfig(".", nil)
	require.NoError(t, err)

	config.ctx = context.TODO()

	// Test the Plan method
	plan, runtimeErr := config.Plan()
	assert.NoError(t, runtimeErr)
	assert.Len(t, plan, 1)
}

func TestPlanSuccess_FileHashRuleSuccess(t *testing.T) {
	expectedContent := "Hello World!"
	// Create a mock HTTP server that returns a specific content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	fs := afero.NewMemMapFs()
	stub := gostub.Stub(&FsFactory, func() afero.Fs { return fs })
	defer stub.Reset()

	sampleConfig := fmt.Sprintf(`  
	data "http" "foo" {  
		url = "%s"  
	}  
  
	rule "file_hash" "bar" {  
		glob = "/testfile"  
		hash = md5(data.http.foo.response_body)  
		algorithm = "md5"  
	}  
	`, server.URL)

	_ = afero.WriteFile(fs, "/testfile", []byte(expectedContent), 0644)
	_ = afero.WriteFile(fs, "test.grept.hcl", []byte(sampleConfig), 0644)

	// Parse the config
	config, err := ParseConfig(".", nil)
	require.NoError(t, err)

	config.ctx = context.TODO()

	plan, err := config.Plan()
	assert.Nil(t, err)
	assert.Empty(t, plan)
}

func dummyFsWithFiles(fileNames []string, contents []string) *gostub.Stubs {
	dummyFs := afero.NewMemMapFs()
	for i, _ := range fileNames {
		_ = afero.WriteFile(dummyFs, fileNames[i], []byte(contents[i]), 0644)
	}

	return gostub.Stub(&FsFactory, func() afero.Fs {
		return dummyFs
	})
}

func TestApplyPlan_multiple_file_fix(t *testing.T) {
	content := `    
	rule "file_hash" sample {    
		glob = "/example/*/testfile"    
		hash = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" // // SHA256 of "hello"    
		algorithm = "sha256"    
	}    
    
	fix "local_file" hello_world{    
		rule_id = rule.file_hash.sample.id  
		paths = rule.file_hash.sample.hash_mismatch_files  
		content = "hello"  
	}    
	`

	stub := dummyFsWithFiles([]string{"test.grept.hcl", "/example/sub1/testfile", "/example/sub2/testfile"}, []string{content, "world", "world"})
	defer stub.Reset()

	config, err := ParseConfig("", nil)
	require.NoError(t, err)

	plan, err := config.Plan()
	require.NoError(t, err)

	err = plan.Apply()
	require.NoError(t, err)

	content1, err := afero.ReadFile(FsFactory(), "/example/sub1/testfile")
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(content1))

	content2, err := afero.ReadFile(FsFactory(), "/example/sub2/testfile")
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(content2))
}
