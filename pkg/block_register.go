package pkg

import (
	"reflect"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type blockConstructor = func(*Config, *hclBlock) block
type blockRegistry map[string]blockConstructor

var baseFactory = map[string]func() any{
	"rule": func() any {
		return new(BaseRule)
	},
	"fix": func() any {
		return new(BaseFix)
	},
	"data": func() any {
		return new(BaseData)
	},
}

func registerFunc(registry blockRegistry, t block) {
	registry[t.Type()] = func(c *Config, hb *hclBlock) block {
		newBlock := reflect.New(reflect.TypeOf(t).Elem()).Elem()
		newBaseBlock := newBaseBlock(c, hb)
		newBaseBlock.setForEach(hb.forEach)
		newBlock.FieldByName("BaseBlock").Set(reflect.ValueOf(newBaseBlock))
		b := newBlock.Addr().Interface().(block)
		if f, ok := baseFactory[t.BlockType()]; ok {
			blockName := cases.Title(language.English).String(t.BlockType())
			newBlock.FieldByName("Base" + blockName).Set(reflect.ValueOf(f()))
		}
		return b
	}
}

var localFactories = make(blockRegistry)

func registerLocal() {
	registerFunc(localFactories, new(LocalBlock))
}

var factories = map[string]blockRegistry{
	"data":  datasourceFactories,
	"rule":  ruleFactories,
	"fix":   fixFactories,
	"local": localFactories,
}
var fixFactories = make(blockRegistry)

func registerFix() {
	registerFunc(fixFactories, new(CopyFileFix))
	registerFunc(fixFactories, new(LocalFileFix))
	registerFunc(fixFactories, new(RenameFileFix))
	registerFunc(fixFactories, new(RmLocalFileFix))
	registerFunc(fixFactories, new(LocalShellFix))
	registerFunc(fixFactories, new(GitIgnoreFix))
	registerFunc(fixFactories, new(YamlTransformFix))
}

var ruleFactories = make(blockRegistry)

func registerRule() {
	registerFunc(ruleFactories, new(FileExistRule))
	registerFunc(ruleFactories, new(FileHashRule))
	registerFunc(ruleFactories, new(MustBeTrueRule))
	registerFunc(ruleFactories, new(DirExistRule))
}

var datasourceFactories = make(blockRegistry)

func registerData() {
	registerFunc(datasourceFactories, new(HttpDatasource))
	registerFunc(datasourceFactories, new(GitIgnoreDatasource))
}
