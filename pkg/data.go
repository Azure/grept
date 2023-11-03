package pkg

type Data interface {
	block
	Load() error
}

type baseData struct {
	*baseBlock
}

func newBaseData(c *Config) baseData {
	return baseData{
		baseBlock: &baseBlock{
			c: c,
		},
	}
}

func (bd baseData) BlockType() string {
	return "data"
}
