package pkg

import "reflect"

type blockConstructor = func(*Config) block
type blockRegistry map[string]blockConstructor

func registerFunc(registry blockRegistry, t block) {
	registry[t.Type()] = func(c *Config) block {
		newBlock := reflect.New(reflect.TypeOf(t).Elem()).Elem()
		newBaseBlock := &BaseBlock{c: c}
		newBlock.FieldByName("BaseBlock").Set(reflect.ValueOf(newBaseBlock))
		return newBlock.Addr().Interface().(block)
	}
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
