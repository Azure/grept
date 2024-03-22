package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/grept/golden"
	"github.com/ahmetb/go-linq/v3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type greptConfigSuite struct {
	suite.Suite
	*testBase
}

func (y *greptConfigSuite) SetupTest() {
	y.testBase = newTestBase()
}

func (y *greptConfigSuite) TearDownTest() {
	y.testBase.teardown()
}

func TestGreptConfigSuite(t *testing.T) {
	suite.Run(t, new(greptConfigSuite))
}

func (s *greptConfigSuite) TestParseConfigHttpBlock() {
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

	t := s.T()
	s.dummyFsWithFiles([]string{"test.Grept.hcl"}, []string{hclConfig})

	// Parse the configuration
	config, err := BuildGreptConfig("", "", nil)
	require.NoError(t, err, "BuildGreptConfig should not return an error")
	_, err = RunGreptPlan(config)
	require.NoError(t, err)
	// ExecuteDuringPlan the parsed configuration
	datas := golden.Blocks[Data](config)
	assert.Len(t, datas, 1, "There should be one data source")

	httpData, ok := datas[0].(*HttpDatasource)
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

func (s *greptConfigSuite) TestPlanError_DatasourceError() {
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
	c, err := BuildGreptConfig("", "", context.TODO())
	require.Nil(t, err)
	_, err = RunGreptPlan(c)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "error making request")
	assert.Contains(t, err.Error(), "data.http.foo")
}

func (s *greptConfigSuite) TestPlanError_FileHashRuleError() {
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
	config, err := BuildGreptConfig("/", "", nil)
	require.NoError(t, err)

	//config.ctx = context.TODO()

	// Test the Plan method
	plan, runtimeErr := RunGreptPlan(config)
	assert.NoError(t, runtimeErr)
	assert.Len(t, plan.FailedRules, 1)
}

func (s *greptConfigSuite) TestPlanSuccess_FileHashRuleSuccess() {
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

	config, err := BuildGreptConfig("/", "", nil)
	require.NoError(t, err)

	//config.ctx = context.TODO()

	plan, err := RunGreptPlan(config)
	assert.Nil(t, err)
	assert.Empty(t, plan.FailedRules)
	assert.Empty(t, plan.Fixes)
}

func (s *greptConfigSuite) TestApplyPlan_multiple_file_fix() {
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

	config, err := BuildGreptConfig("/", "", nil)
	require.NoError(t, err)

	plan, err := RunGreptPlan(config)
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

func (s *greptConfigSuite) TestConfig_MultipleTypeRules() {
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
	c, err := BuildGreptConfig("", "/", context.TODO())
	assert.NoError(t, err)
	_, err = RunGreptPlan(c)
	require.NoError(t, err)
	rules := golden.Blocks[Rule](c)
	assert.Len(t, rules, 2)
	var types []string
	linq.From(rules).Select(func(i interface{}) interface{} {
		return i.(Rule).Type()
	}).ToSlice(&types)
	assert.Contains(t, types, "file_hash")
	assert.Contains(t, types, "must_be_true")
}

func (s *greptConfigSuite) TestHttpDatasource_DefaultMethodShouldBeGet() {
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

			s.dummyFsWithFiles([]string{"/test.grept.hcl"}, []string{hclConfig})
			config, err := BuildGreptConfig("", "/", context.TODO())
			s.NoError(err)
			_, err = RunGreptPlan(config)
			s.NoError(err)
			d := golden.Blocks[Data](config)[0].(*HttpDatasource)
			assert.Equal(s.T(), c.want, d.Method)
		})
	}
}

func (s *greptConfigSuite) TestAnyRuleFailShouldTriggerFix() {
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

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})

	// Parse the configuration
	config, err := BuildGreptConfig("", "", nil)
	s.NoError(err)
	plan, err := RunGreptPlan(config)
	s.NoError(err)
	err = plan.Apply()
	s.NoError(err)
	exists, err := afero.Exists(s.fs, "/file")
	s.NoError(err)
	s.True(exists)
}

func (s *greptConfigSuite) TestMultipleRulesTriggerSameFixShouldExecuteOnlyOnce() {
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

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})

	// Parse the configuration
	config, err := BuildGreptConfig("", "", nil)
	s.NoError(err)
	plan, err := RunGreptPlan(config)
	s.NoError(err)
	s.Len(plan.Fixes, 1)
}

func (s *greptConfigSuite) TestEmptyConfigFolderShouldThrowError() {
	_, err := BuildGreptConfig("/", "/", context.TODO())
	s.NotNil(err)
	s.Contains(err.Error(), "no `.grept.hcl` file found")
}

func (s *greptConfigSuite) TestParseConfigBeforePlan_UnknownValueShouldNotTriggerError() {
	t := s.T()
	expectedContent := "{\"hello\": \"world\"}"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	sampleConfig := fmt.Sprintf(`
	data "http" "foo" {
		url = "%s"
	}

	rule "must_be_true" "bar" {
		condition = yamldecode(data.http.foo.response_body).hello == "world"
	}
	`, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	_, err := BuildGreptConfig("/", "", nil)
	require.NoError(t, err)
}

func (s *greptConfigSuite) TestLocalWithRuleAndFix() {
	hcl := `
	locals {
		path = "LICENSE"
	}

	rule "file_hash" sample {
		glob = local.path
		hash = "abc123"
		algorithm = "sha256"
	}

	fix "local_file" hello_world{
		rule_ids = [rule.file_hash.sample.id]
		paths = [local.path]
		content = "Hello, world!"
	}
`
	t := s.T()
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hcl})
	config, err := BuildGreptConfig("", "", nil)
	require.NoError(t, err)
	_, err = RunGreptPlan(config)
	require.NoError(t, err)
	fixes := golden.Blocks[Fix](config)
	require.Len(t, fixes, 1)
	fix := fixes[0].(*LocalFileFix)
	assert.Equal(t, "LICENSE", fix.Paths[0])
}

func (s *greptConfigSuite) TestLocalBetweenDataAndRule() {
	expectedContent := "world"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	sampleConfig := fmt.Sprintf(`
	locals{
		content = data.http.foo.response_body
	}

	data "http" "foo" {
		url = "%s"
	}

	rule "must_be_true" "bar" {
		condition = local.content == "world"
	}
	`, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	c, err := BuildGreptConfig("/", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(c)
	s.NoError(err)
	rules := golden.Blocks[Rule](c)
	s.True(rules[0].(*MustBeTrueRule).Condition)
}

func (s *greptConfigSuite) TestForEach_from_data_to_fix() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Parse the JSON body
		var body map[string]interface{}
		err = json.Unmarshal(bodyBytes, &body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Get the 'query' value
		query, ok := body["query"].(string)
		if !ok {
			http.Error(w, "Invalid 'query' value", http.StatusBadRequest)
			return
		}
		// Write the 'query' value as the response
		_, _ = w.Write([]byte(query))
	}))
	defer server.Close()
	hclConfig := fmt.Sprintf(`
    locals {
        items = toset(["item1", "item2", "item3"])
    }

	data "http" echo {
		for_each = local.items
		url = "%s"
		request_body = jsonencode({
    		query = each.value
  		})
	}

    rule "must_be_true" sample {
        for_each = local.items
        condition = each.value != data.http.echo[each.value].response_body
    }

    fix "local_file" hello_world{
		for_each = local.items
        rule_ids = [rule.must_be_true.sample[each.value].id]
        paths = [each.value]
        content = each.value
    }
    `, server.URL)

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{hclConfig})

	config, err := BuildGreptConfig("", "", nil)
	require.NoError(s.T(), err)

	p, err := RunGreptPlan(config)
	require.NoError(s.T(), err)
	err = p.Apply()
	require.NoError(s.T(), err)

	// Verify that the file has been created successfully
	exists, err := afero.Exists(s.fs, "/items1")
	s.NoError(err)
	s.False(exists)
}

func (s *greptConfigSuite) TestPlanOnlyAddFixWhenCheckErrNotNil() {
	t := s.T()
	content := `
	rule "must_be_true" sample {
		condition = true
	}

	fix "local_file" hello_world{
		rule_ids = [rule.must_be_true.sample.id]
		paths = ["/path/to/file.txt"]
		content = "Hello, world!"
	}
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := BuildGreptConfig("", "", nil)
	require.NoError(t, err)
	plan, err := RunGreptPlan(config)
	require.NoError(t, err)
	s.Len(plan.Fixes, 0)
}

func (s *greptConfigSuite) TestApplyPlan_file_fix_with_null_mode() {
	t := s.T()
	content := `
	rule "file_hash" "sample" {
		glob = "/example/testfile"
		hash = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" // SHA256 of "hello"
		algorithm = "sha256"
	}

	fix "local_file" "hello_world" {
		rule_ids = [rule.file_hash.sample.id]
		paths = rule.file_hash.sample.hash_mismatch_files
		content = "hello"
		mode = null
	}
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl", "/example/testfile"}, []string{content, "world"})

	config, err := BuildGreptConfig("/", "", nil)
	require.NoError(t, err)

	plan, err := RunGreptPlan(config)
	require.NoError(t, err)

	err = plan.Apply()
	require.NoError(t, err)

	content1, err := afero.ReadFile(FsFactory(), "/example/testfile")
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(content1))
}
