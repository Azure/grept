package pkg

type Data interface {
	block
	Load() error
}

type baseData struct {
	*baseBlock
}

func (bd baseData) BlockType() string {
	return "data"
}
