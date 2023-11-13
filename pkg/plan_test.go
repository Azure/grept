package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlan_String(t *testing.T) {
	tests := []struct {
		name string
		plan *Plan
		want []string
	}{
		{
			name: "Test Plan String",
			plan: &Plan{
				FailedRules: []*FailedRule{
					{
						Rule: &FileHashRule{
							BaseBlock: &BaseBlock{
								name: "test-rule",
								id:   "1",
							},
							Glob:      "test-glob",
							Hash:      "test-hash",
							Algorithm: "sha1",
						},
						CheckError: fmt.Errorf("test error"),
					},
				},
				Fixes: map[string]Fix{
					"test_id": &LocalFileFix{
						BaseBlock: &BaseBlock{
							name: "test-fix",
						},
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
