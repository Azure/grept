package pkg

type Data interface {
	block
}

type baseData struct{}

func (bd baseData) BlockType() string {
	return "data"
}
