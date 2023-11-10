package pkg

type Fix interface {
	block
	ApplyFix() error
	GetRuleIds() []string
}

type baseFix struct{}

func (bf baseFix) BlockType() string {
	return "fix"
}
