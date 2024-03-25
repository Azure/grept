package golden

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
)

type directedAcyclicGraph interface {
	GetVertices() map[string]interface{}
	GetAncestors(id string) (map[string]interface{}, error)
	GetChildren(id string) (map[string]interface{}, error)
	buildDag(blocks []Block) error
	runDag(onReady func(Block) error) error
}

type Config interface {
	directedAcyclicGraph
	Context() context.Context
	EvalContext() *hcl.EvalContext
	RunPrePlan() error
	RunPlan() error
	expandBlock(b Block) ([]Block, error)
}

func Blocks[T Block](c directedAcyclicGraph) []T {
	var r []T
	for _, b := range c.GetVertices() {
		t, ok := b.(T)
		if ok {
			r = append(r, t)
		}
	}
	return r
}

func InitConfig(config Config, hclBlocks []*HclBlock) error {
	var err error

	var blocks []Block
	for _, hb := range hclBlocks {
		b, wrapError := wrapBlock(config, hb)
		if wrapError != nil {
			err = multierror.Append(wrapError)
			continue
		}
		blocks = append(blocks, b)
	}
	if err != nil {
		return err
	}
	// If there's dag error, return dag error first.
	err = config.buildDag(blocks)
	if err != nil {
		return err
	}
	err = config.RunPrePlan()
	if err != nil {
		return err
	}

	return nil
}

func wrapBlock(c Config, hb *HclBlock) (Block, error) {
	blockFactories := factories[hb.Type]
	blockType := ""
	if len(hb.Labels) > 0 {
		blockType = hb.Labels[0]
	}
	f, ok := blockFactories[blockType]
	if !ok {
		return nil, fmt.Errorf("unregistered %s: %s", hb.Type, blockType)
	}
	return f(c, hb), nil
}

func blocks(c directedAcyclicGraph) []Block {
	var blocks []Block
	for _, n := range c.GetVertices() {
		blocks = append(blocks, n.(Block))
	}
	return blocks
}

func castBlock[T Block](s []Block) []T {
	var r []T
	for _, b := range s {
		r = append(r, b.(T))
	}
	return r
}
