package pkg

import "github.com/heimdalr/dag"

type Dag struct {
	*dag.DAG
}

func newDag() *Dag {
	return &Dag{
		DAG: dag.NewDAG(),
	}
}
