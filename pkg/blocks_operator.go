package pkg

import "sync"

type BlocksOperator struct {
	blocks []block
	wg     sync.WaitGroup
}

func (o *BlocksOperator) addBlock(b block) {
	o.blocks = append(o.blocks, b)
	o.wg.Add(1)
}
