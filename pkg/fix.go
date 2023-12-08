package pkg

type Fix interface {
	block
	GetRuleIds() []string
}

type baseFix struct{}

func (bf baseFix) BlockType() string {
	return "fix"
}
