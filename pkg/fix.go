package pkg

type Fix interface {
	block
	ApplyFix() error
	GetRuleId() string
}

type baseFix struct{}

func (bf baseFix) BlockType() string {
	return "fix"
}
