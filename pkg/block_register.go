package pkg

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"reflect"
)

type blockConstructor = func(*Config, *hclsyntax.Block) block
type blockRegistry map[string]blockConstructor

func registerFunc(registry blockRegistry, t block) {
	registry[t.Type()] = func(c *Config, hb *hclsyntax.Block) block {
		newBlock := reflect.New(reflect.TypeOf(t).Elem()).Elem()
		newBaseBlock := &BaseBlock{
			c:  c,
			hb: hb,
		}
		newBlock.FieldByName("BaseBlock").Set(reflect.ValueOf(newBaseBlock))
		return newBlock.Addr().Interface().(block)
	}
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
