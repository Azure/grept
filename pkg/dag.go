package pkg

import (
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/heimdalr/dag"
)

type Dag struct {
	*dag.DAG
}

func newDag(blocks []block) (*Dag, error) {
	g := &Dag{
		DAG: dag.NewDAG(),
	}
	var walkErr error
	for _, b := range blocks {
		err := g.AddVertexByID(blockAddress(b.HclSyntaxBlock()), b)
		if err != nil {
			walkErr = multierror.Append(walkErr, err)
		}
	}
	for _, b := range blocks {
		diag := hclsyntax.Walk(b.HclSyntaxBlock().Body, dagWalker{dag: g, rootBlock: b})
		if diag.HasErrors() {
			walkErr = multierror.Append(walkErr, diag.Errs()...)
		}
	}
	return g, walkErr
}
