package fixes

func init() {
	register("local_file", func() Fix { return &LocalFile{} })
}

type Fix interface {
	ApplyFix() error
}

var FixFactories = map[string]func() Fix{}

func register(name string, factory func() Fix) {
	FixFactories[name] = factory
}
