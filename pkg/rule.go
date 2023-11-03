package pkg

type Rule interface {
	block
	Check() (checkError error, runtimeError error)
}

type baseRule struct {
	*baseBlock
}

func (br baseRule) BlockType() string {
	return "rule"
}
