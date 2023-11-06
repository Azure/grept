package pkg

type Rule interface {
	block
	Check() (checkError error, runtimeError error)
}

type baseRule struct{}

func (br baseRule) BlockType() string {
	return "rule"
}
