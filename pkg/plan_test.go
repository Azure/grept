package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclwrite"
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
