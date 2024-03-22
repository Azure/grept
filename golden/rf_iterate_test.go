package golden

import (
	"github.com/hashicorp/hcl/v2"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
)

func TestRefIterator(t *testing.T) {
	tests := []struct {
		name string
		hcl  string
		want string
	}{
		{
			name: "data ref iterator",
			hcl:  `data.source.attribute`,
			want: "data.source.attribute",
		},
		{
			name: "resource ref iterator",
			hcl:  `resource.dummy.id`,
			want: "resource.dummy.id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _ := hclsyntax.ParseExpression([]byte(tt.hcl), "", hcl.InitialPos)
			ts := expr.Variables()

			root := name(ts[0][0])
			iterator := refIters[root]
			got := iterator(ts[0], 0)

			assert.Equal(t, tt.want, got[0])
		})
	}
}
