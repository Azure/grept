package pkg

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type configSuite struct {
	suite.Suite
	*testBase
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(configSuite))
}

func (s *configSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *configSuite) TearDownTest() {
	s.teardown()
}

func (s *configSuite) TestParseConfig() {
	content := `  
	rule "file_hash" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_ids = [rule.file_hash.sample.id]
		paths = ["/path/to/file.txt"]  
		content = "Hello, world!"
	}  
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	t := s.T()

	config, err := ParseConfig("", nil)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(config.Rules))
	fhr, ok := config.Rules[0].(*FileHashRule)
	require.True(t, ok)
	assert.Equal(t, "*.txt", fhr.Glob)
	assert.Equal(t, "abc123", fhr.Hash)
	assert.Equal(t, "sha256", fhr.Algorithm)

	assert.Equal(t, 1, len(config.Fixes))
	lff, ok := config.Fixes[0].(*LocalFileFix)
	require.True(t, ok)
	assert.Regexp(t, `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`, lff.RuleIds[0])
	assert.Equal(t, "/path/to/file.txt", lff.Paths[0])
	assert.Equal(t, "Hello, world!", lff.Content)
}

func (s *configSuite) TestUnregisteredFix() {
	hcl := `  
	fix "unregistered_fix" sample {  
		rule_id = "c01d7cf6-ec3f-47f0-9556-a5d6e9009a43"  
		path = "/path/to/file.txt"  
		content = "Hello, world!"  
	}  
	`

	t := s.T()
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	_, err := ParseConfig(".", nil)
	require.NotNil(t, err)
	expectedError := "unregistered fix: unregistered_fix"
	assert.Contains(t, err.Error(), expectedError)
}

func (s *configSuite) TestUnregisteredRule() {
	hcl := `  
	rule "unregistered_rule" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
	`

	t := s.T()
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	_, err := ParseConfig(".", nil)
	require.NotNil(t, err)

	expectedError := "unregistered rule: unregistered_rule"
	assert.Contains(t, err.Error(), expectedError)
}

func (s *configSuite) TestInvalidBlockType() {
	hcl := `  
	invalid_block "invalid_type" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
	`

	t := s.T()
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	_, err := ParseConfig("", nil)
	require.NotNil(t, err)

	expectedError := "invalid block type: invalid_block"
	assert.Contains(t, err.Error(), expectedError)
}

func (s *configSuite) TestEvalContextRef() {
	hcl := `
	rule "file_hash" sample {  
		glob = "LICENSE"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_ids = [rule.file_hash.sample.id]
		paths = [rule.file_hash.sample.glob]  
		content = "Hello, world!"
	}
`
	t := s.T()
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	config, err := ParseConfig("", nil)
	assert.NoError(t, err)
	require.Equal(t, 1, len(config.Fixes))
	fix := config.Fixes[0].(*LocalFileFix)
	assert.Equal(t, "LICENSE", fix.Paths[0])
}

func (s *configSuite) TestFunctionInEvalContext() {
	t := s.T()
	fileContent := "Hello, world!"
	configStr := fmt.Sprintf(`  
	rule "file_hash" "test_rule" {  
		glob = "/testfile"  
		hash = md5("%s")  
		algorithm = "md5"  
	}  
	`, fileContent)
	s.dummyFsWithFiles([]string{"/testfile", "test.grept.hcl"}, []string{fileContent, configStr})

	config, err := ParseConfig(".", nil)
	require.NoError(t, err)
	require.Equal(t, 1, len(config.Rules))
	rule, ok := config.Rules[0].(*FileHashRule)
	require.True(t, ok)
	_, err = rule.Check()
	assert.NoError(t, err)
}

func (s *configSuite) TestParseConfigHttpBlock() {
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
	t := s.T()
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})

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

func (s *configSuite) TestPlanError_DatasourceError() {
	t := s.T()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Mock server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`  
	data "http" "foo" {  
		url = "%s"
		retry_max = 0
	}  
`, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})
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

func (s *configSuite) TestPlanError_FileHashRuleError() {
	t := s.T()
	// Create a mock HTTP server that returns a specific content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

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
	s.dummyFsWithFiles([]string{"/testfile", "test.grept.hcl"}, []string{"Different content", sampleConfig})
	// Parse the config
	config, err := ParseConfig(".", nil)
	require.NoError(t, err)

	config.ctx = context.TODO()

	// Test the Plan method
	plan, runtimeErr := config.Plan()
	assert.NoError(t, runtimeErr)
	assert.Len(t, plan.FailedRules, 1)
}

func (s *configSuite) TestPlanSuccess_FileHashRuleSuccess() {
	t := s.T()
	expectedContent := "Hello World!"
	// Create a mock HTTP server that returns a specific content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(expectedContent))
	}))
	defer server.Close()

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
	s.dummyFsWithFiles([]string{"/testfile", "test.grept.hcl"}, []string{expectedContent, sampleConfig})

	config, err := ParseConfig(".", nil)
	require.NoError(t, err)

	config.ctx = context.TODO()

	plan, err := config.Plan()
	assert.Nil(t, err)
	assert.Empty(t, plan.FailedRules)
	assert.Empty(t, plan.Fixes)
}

func (s *configSuite) TestApplyPlan_multiple_file_fix() {
	t := s.T()
	content := `    
	rule "file_hash" sample {    
		glob = "/example/*/testfile"    
		hash = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" // SHA256 of "hello"    
		algorithm = "sha256"    
	}    
    
	fix "local_file" hello_world{    
		rule_ids = [rule.file_hash.sample.id]  
		paths = rule.file_hash.sample.hash_mismatch_files  
		content = "hello"  
	}    
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl", "/example/sub1/testfile", "/example/sub2/testfile"}, []string{content, "world", "world"})

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

func (s *configSuite) TestConfig_MultipleTypeRules() {
	t := s.T()
	hcl := `
rule file_hash license {
  glob = "LICENSE"
  hash = sha1("this is a fake license")
}

rule must_be_true test {
  condition = env("OS") == "windows"
}
`

	s.dummyFsWithFiles([]string{"/test.grept.hcl"}, []string{hcl})
	c, err := ParseConfig("/", context.TODO())
	assert.NoError(t, err)
	assert.Len(t, c.Rules, 2)
	assert.Equal(t, "file_hash", c.Rules[0].Type())
	assert.Equal(t, "must_be_true", c.Rules[1].Type())
}

func (s *configSuite) TestHttpDatasource_DefaultMethodShouldBeGet() {
	cases := []struct {
		desc   string
		method string
		want   string
	}{
		{
			desc:   "Should apply default value",
			method: "",
			want:   "GET",
		},
		{
			desc:   "User's input is as same as the default value",
			method: "GET",
			want:   "GET",
		},
		{
			desc:   "User's input should take precedence over default value",
			method: "POST",
			want:   "POST",
		},
	}
	for _, c := range cases {
		s.Run(c.desc, func() {
			assignment := fmt.Sprintf("method = \"%s\"", c.method)
			if c.method == "" {
				assignment = ""
			}
			hclConfig := fmt.Sprintf(`  
	data "http" "example" {  
		url = "http://example.com"  
		request_body = "Hello" 
		%s
		request_headers = {  
			"Content-Type" = "application/json"  
			"Accept" = "application/json"  
		}  
	}  
	`, assignment)

			config, diag := hclsyntax.ParseConfig([]byte(hclConfig), "test.grept.hcl", hcl.InitialPos)
			require.False(s.T(), diag.HasErrors())
			h := &HttpDatasource{
				BaseBlock: &BaseBlock{
					c: &Config{
						ctx: context.TODO(),
					},
					hb: config.Body.(*hclsyntax.Body).Blocks[0],
				},
			}
			err := eval(h)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), c.want, h.Method)
		})
	}
}

func (s *configSuite) TestAnyRuleFailShouldTriggerFix() {
	hclConfig := `  
	rule "must_be_true" true {
		condition = true
	}
	rule "must_be_true" false {
		condition = false
	}
	fix "local_file" file {
		rule_ids = [rule.must_be_true.true.id, rule.must_be_true.false.id]
		paths = ["/file"]
		content = ""
	}
	`

	dir := "."
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})

	// Parse the configuration
	config, err := ParseConfig(dir, nil)
	s.NoError(err)
	plan, err := config.Plan()
	s.NoError(err)
	err = plan.Apply()
	s.NoError(err)
	exists, err := afero.Exists(s.fs, "/file")
	s.NoError(err)
	s.True(exists)
}

func (s *configSuite) TestMultipleRulesTriggerSameFixShouldExecuteOnlyOnce() {
	hclConfig := `  
	rule "must_be_true" one {
		condition = false
	}
	rule "must_be_true" two {
		condition = false
	}
	fix "local_file" file {
		rule_ids = [rule.must_be_true.one.id, rule.must_be_true.two.id]
		paths = ["/file"]
		content = ""
	}
	`

	dir := "."
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})

	// Parse the configuration
	config, err := ParseConfig(dir, nil)
	s.NoError(err)
	plan, err := config.Plan()
	s.NoError(err)
	s.Len(plan.Fixes, 1)
}
