package pkg

import (
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
)

type BlocksOperator struct {
	c      *Config
	blocks sets.Set
}

func NewBlocksOperator(c *Config) *BlocksOperator {
	return &BlocksOperator{
		c:      c,
		blocks: hashset.New(),
	}
}

func (o *BlocksOperator) addBlock(b block) {
	o.blocks.Add(b)
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
