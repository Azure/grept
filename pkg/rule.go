package pkg

import "reflect"

type Rule interface {
	block
	Check() error
	CheckError() error
}

type BaseRule struct {
	checkErr error
}

func (br BaseRule) BlockType() string {
	return "rule"
}

func (br BaseRule) CheckError() error {
	return br.checkErr
}

func logCheckError[T Rule](rule T, checkErr error) {
	newBaseRule := BaseRule{
		checkErr: checkErr,
	}
	baseRuleField := reflect.ValueOf(rule).Elem().FieldByName("BaseRule")
	baseRuleField.Set(reflect.ValueOf(newBaseRule))
}
