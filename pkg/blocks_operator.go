package pkg

import "sync"

type BlocksOperator struct {
	c      *Config
	blocks []block
	wg     sync.WaitGroup
}

func NewBlocksOperator(c *Config) *BlocksOperator {
	return &BlocksOperator{
		c: c,
	}
}

func (o *BlocksOperator) addBlock(b block) {
	o.blocks = append(o.blocks, b)
	o.wg.Add(1)
}

func (o *BlocksOperator) notifyOnExecuted(b block, success bool) {
	for _, next := range b.getDownstreams() {
		next.notifyOnExecuted(b, success)
	}
	o.wg.Done()
}

func (o *BlocksOperator) blocksCount() int {
	return len(o.blocks)
}
