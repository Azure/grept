package pkg

import (
	"github.com/emirpasic/gods/sets/hashset"
	"sync"

	"github.com/emirpasic/gods/sets"
)

type BlocksOperator struct {
	c      *Config
	blocks sets.Set
	wg     sync.WaitGroup
}

func NewBlocksOperator(c *Config) *BlocksOperator {
	return &BlocksOperator{
		c:      c,
		blocks: hashset.New(),
	}
}

func (o *BlocksOperator) addBlock(b block) {
	o.blocks.Add(b)
	o.wg.Add(1)
}

func (o *BlocksOperator) notifyOnExecuted(b block, success bool) {
	for _, next := range b.getDownstreams() {
		next.notifyOnExecuted(b, success)
	}
	o.wg.Done()
}

func (o *BlocksOperator) blocksCount() int {
	return o.blocks.Size()
}

func (o *BlocksOperator) Blocks() []block {
	var blocks []block
	for _, v := range o.blocks.Values() {
		blocks = append(blocks, v.(block))
	}
	return blocks
}
