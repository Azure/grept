package rules

func init() {
	register("file_hash", func() Rule { return &FileHashRule{} })
}

type Rule interface {
	Check() error
	Validate() error
}

var RuleFactories = map[string]func() Rule{}

func register(name string, factory func() Rule) {
	RuleFactories[name] = factory
}
