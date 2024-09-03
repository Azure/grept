package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/Azure/golden"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
)

func TestPlan_String(t *testing.T) {
	tests := []struct {
		name string
		plan *GreptPlan
		want []string
	}{
		{
			name: "Test Plan String",
			plan: &GreptPlan{
				FailedRules: []*FailedRule{
					{
						Rule: &FileHashRule{
							BaseBlock: golden.NewBaseBlock(nil, golden.NewHclBlock(&hclsyntax.Block{
								Type:   "rule",
								Labels: []string{"file_hash", "test-rule"},
								Body:   &hclsyntax.Body{},
							}, hclwrite.NewBlock("rule", []string{"file_hash", "test-rule"}), nil)),
							Glob:      "test-glob",
							Hash:      "test-hash",
							Algorithm: "sha1",
						},
						CheckError: fmt.Errorf("test error"),
					},
				},
				Fixes: map[string]Fix{
					"test_id": &LocalFileFix{
						BaseBlock: golden.NewBaseBlock(nil, golden.NewHclBlock(&hclsyntax.Block{
							Type:   "fix",
							Labels: []string{"local_file", "test-fix"},
							Body:   &hclsyntax.Body{},
						}, hclwrite.NewBlock("fix", []string{"local_file", "test-fix"}), nil)),
						Paths:   []string{"test-path"},
						Content: "test-content",
					},
				},
			},
			want: []string{
				"rule.file_hash.test-rule",
				"fix.local_file.test-fix would be apply",
				"\"test-path\"",
				"\"test-content\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.plan
			got := p.String()
			for _, w := range tt.want {
				assert.Contains(t, got, w)
			}
		})
	}
}

func (s *greptConfigSuite) TestUnTriggeredFixShouldNotBeApplied() {
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
	config, err := BuildGreptConfig("", "", nil, nil)
	require.NoError(t, err)
	plan, err := RunGreptPlan(config)
	require.NoError(t, err)

	// Apply the plan
	err = plan.Apply()
	require.NoError(t, err)

	// Verify that the file has not been created
	exists, err := afero.Exists(s.fs, "/path/to/file.txt")
	s.NoError(err)
	s.False(exists)
}
