package golden

import (
	"github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/heimdalr/dag"
)

type Dag struct {
	*dag.DAG
}

func newDag() *Dag {
	return &Dag{
		DAG: dag.NewDAG(),
	}
}

func (d *Dag) buildDag(blocks []Block) error {
	var walkErr error
	for _, b := range blocks {
		err := d.AddVertexByID(b.Address(), b)
		if err != nil {
			walkErr = multierror.Append(walkErr, err)
		}
	}
	for _, b := range blocks {
		diag := hclsyntax.Walk(b.HclBlock().Body, newDagWalker(d, b.Address()))
		if diag.HasErrors() {
			walkErr = multierror.Append(walkErr, diag.Errs()...)
		}
	}
	return walkErr
}

func (d *Dag) addEdge(from, to string) error {
	err := d.AddEdge(from, to)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dag) runDag(c Config, onReady func(Block) error) error {
	var err error
	pending := linkedlistqueue.New()
	for _, n := range d.GetRoots() {
		pending.Enqueue(n.(Block))
	}
	for !pending.Empty() {
		next, _ := pending.Dequeue()
		b := next.(Block)
		// the node has already been expandable and deleted from dag
		address := b.Address()
		exist := d.exist(address)
		if !exist {
			continue
		}
		ancestors, dagErr := d.GetAncestors(address)
		if dagErr != nil {
			return dagErr
		}
		ready := true
		for upstreamAddress := range ancestors {
			v, dagErr := d.GetVertex(upstreamAddress)
			if dagErr != nil {
				return dagErr
			}
			if !v.(Block).isReadyForRead() {
				ready = false
				break
			}
		}
		if !ready {
			continue
		}
		if b.expandable() {
			expandedBlocks, err := c.expandBlock(b)
			if err != nil {
				return err
			}
			newPending := linkedlistqueue.New()
			for _, eb := range expandedBlocks {
				newPending.Enqueue(eb)
			}
			for _, b := range pending.Values() {
				newPending.Enqueue(b)
			}
			pending = newPending
			continue
		}
		if callbackErr := onReady(b); callbackErr != nil {
			err = multierror.Append(err, callbackErr)
		}
		// this address might be expandable during onReady and no more exist.
		exist = d.exist(address)
		if !exist {
			continue
		}
		children, dagErr := d.GetChildren(address)
		if dagErr != nil {
			return dagErr
		}
		for _, n := range children {
			pending.Enqueue(n)
		}
	}
	return err
}

func (d *Dag) exist(address string) bool {
	n, existErr := d.GetVertex(address)
	notExist := n == nil || existErr != nil
	return !notExist
}
