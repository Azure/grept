package pkg

type blockConstructor = func(*Config) block
type blockRegistry map[string]blockConstructor

func registerFunc(registry blockRegistry, t block) {
	registry[t.Type()] = t.constructor()
}

var fixFactories = make(blockRegistry)

func registerFix() {
	registerFunc(fixFactories, new(LocalFileFix))
	registerFunc(fixFactories, new(RenameFileFix))
	registerFunc(fixFactories, new(RmLocalFileFix))
}

var ruleFactories = map[string]func(*Config) block{}

func registerRule() {
	registerFunc(ruleFactories, new(FileHashRule))
	registerFunc(ruleFactories, new(MustBeTrueRule))
	registerFunc(ruleFactories, new(DirExistRule))
}

var datasourceFactories = map[string]func(*Config) block{}

func registerData() {
	registerFunc(datasourceFactories, new(HttpDatasource))
	registerFunc(datasourceFactories, new(GitIgnoreDatasource))
}
