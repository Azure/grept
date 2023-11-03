package pkg

type Fix interface {
	block
	ApplyFix() error
	GetRuleId() string
}

type baseFix struct {
	*baseBlock
}

func (bf baseFix) BlockType() string {
	return "fix"
}

func newBaseFix(c *Config) baseFix {
	return baseFix{
		baseBlock: &baseBlock{c: c},
	}
}
