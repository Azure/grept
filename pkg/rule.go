package pkg

type Rule interface {
	block
	CheckError() error
	setCheckError(error)
}

type BaseRule struct {
	checkErr error
}

func (br *BaseRule) BlockType() string {
	return "rule"
}

func (br *BaseRule) CheckError() error {
	return br.checkErr
}

func (br *BaseRule) setCheckError(err error) {
	br.checkErr = err
}
