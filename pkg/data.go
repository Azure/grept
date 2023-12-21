package pkg

type Data interface {
	block
}

type BaseData struct{}

func (bd *BaseData) BlockType() string {
	return "data"
}
