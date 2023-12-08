package pkg

type Fix interface {
	block
	GetRuleIds() []string
	Apply() error
}

type baseFix struct{}

func (bf baseFix) BlockType() string {
	return "fix"
}
