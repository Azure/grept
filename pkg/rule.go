package pkg

type Rule interface {
	block
	Check() (checkError error, runtimeError error)
}

type baseRule struct {
	*baseBlock
}

func newBaseRule(c *Config) baseRule {
	return baseRule{
		baseBlock: &baseBlock{
			c: c,
		},
	}
}

func (br baseRule) BlockType() string {
	return "rule"
}
