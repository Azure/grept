package pkg

import (
	"fmt"

	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/heimdalr/dag"
)

var _ hclsyntax.Walker = dagWalker{}
var keywords sets.Set = hashset.New("data", "fix", "rule")

type dagWalker struct {
	dag       *dag.DAG
	rootBlock block
}

func (d dagWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	diag := hcl.Diagnostics{}
	if expr, ok := node.(hclsyntax.Expression); ok {
		traversals := expr.Variables()
		for _, traversal := range traversals {
			for i, traverser := range traversal {
				refIter, ok := refIters[name(traverser)]
				if !ok {
					continue
				}
				if ref := refIter(traversal, i); ref != nil {
					src := *ref
					dest := blockAddress(d.rootBlock)
					children, err := d.dag.GetChildren(src)
					if err != nil {
						diag = diag.Append(&hcl.Diagnostic{
							Severity: hcl.DiagError,
							Summary:  fmt.Sprintf("cannot get children from %s", src),
							Detail:   err.Error(),
						})
						continue
					}
					if _, ok := children[dest]; !ok {
						err := d.dag.AddEdge(src, dest)
						if err != nil {
							diag = diag.Append(&hcl.Diagnostic{
								Severity: hcl.DiagError,
								Summary:  "cannot add edge",
								Detail:   err.Error(),
							})
						}
					}
				}
			}
		}
	}
	return diag
}

func (d dagWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	return nil
}
