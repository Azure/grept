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

func (o *BlocksOperator) notifyOnEvaluated(b block) {
	o.wg.Done()
}
