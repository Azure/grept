package rules

type Rule interface {
	Check() error
	Validate() error
}
