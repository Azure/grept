package pkg

import (
	"fmt"
	"github.com/emirpasic/gods/sets"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
)

var validBlockTypes sets.Set = hashset.New("data", "rule", "fix")

type parserFactory func(c *Config) func(*hclsyntax.Block) error

var ruleParser parserFactory = func(c *Config) func(*hclsyntax.Block) error {
	return c.parseFunc("rule", ruleFactories, func(cc *Config, b block) {
		cc.Rules = append(cc.Rules, b.(Rule))
	})

}
var fixParser parserFactory = func(c *Config) func(*hclsyntax.Block) error {
	return c.parseFunc("fix", fixFactories, func(cc *Config, b block) {
		cc.Fixes = append(cc.Fixes, b.(Fix))
	})
}
var dataParser parserFactory = func(c *Config) func(*hclsyntax.Block) error {
	return c.parseFunc("data", datasourceFactories, func(cc *Config, b block) {
		cc.DataSources = append(cc.DataSources, b.(Data))
	})
}

type parsers []parserFactory

var blockParsers parsers = []parserFactory{
	dataParser,
	ruleParser,
	fixParser,
}

type Datas []Data
type Rules []Rule
type Fixes []Fix

func Values[T block](slice []T) cty.Value {
	if len(slice) == 0 {
		return cty.EmptyObjectVal
	}
	res := map[string]cty.Value{}
	valuesMap := map[string]map[string]cty.Value{}

	for _, r := range slice {
		inner := valuesMap[r.Type()]
		if inner == nil {
			inner = map[string]cty.Value{}
		}
		inner[r.Name()] = r.Value()
		res[r.Type()] = cty.MapVal(inner)
		valuesMap[r.Type()] = inner
	}
	return cty.MapVal(res)
}

type Config struct {
	basedir     string
	DataSources Datas
	Rules       Rules
	Fixes       Fixes
}

func (c *Config) EvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Functions: hcl2template.Functions(c.basedir),
		Variables: map[string]cty.Value{
			"data": Values(c.DataSources),
			"rule": Values(c.Rules),
		},
	}
}

func (c *Config) parseFunc(expectedBlockType string, factories map[string]func(*Config) block, postParseFunc func(*Config, block)) func(*hclsyntax.Block) error {
	return func(hb *hclsyntax.Block) error {
		if hb.Type != expectedBlockType {
			return nil
		}
		if len(hb.Labels) != 2 {
			return fmt.Errorf("invalid labels for rule %s, expect labels with length 2 (%s)", concatLabels(hb.Labels), hb.Range().String())
		}
		t := hb.Labels[0]
		f, ok := factories[t]
		if !ok {
			return fmt.Errorf("unregistered %s: %s, %s", expectedBlockType, t, hb.Range().String())
		}
		b := f(c)
		err := b.Parse(hb)
		if err != nil {
			return err
		}
		postParseFunc(c, b)
		return nil
	}
}

func ParseConfig(dir, filename, content string) (*Config, error) {
	config := &Config{
		basedir: dir,
	}

	file, diag := hclsyntax.ParseConfig([]byte(content), filename, hcl.InitialPos)
	if diag.HasErrors() {
		return nil, diag
	}
	body := file.Body.(*hclsyntax.Body)
	var err error
	// First loop: parse all rule blocks
	for _, b := range body.Blocks {
		if !validBlockTypes.Contains(b.Type) {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", b.Type, b.Range().String()))
			continue
		}
	}
	for _, parser := range blockParsers {
		for _, b := range body.Blocks {
			parseError := parser(config)(b)
			if parseError != nil {
				err = multierror.Append(err, parseError)
			}
		}
	}

	return config, err
}

func ApplyRulesAndFixes(config *Config) error {
	for _, rule := range config.Rules {
		err := rule.Check()
		if err != nil {
			// If a rule check fails, apply the corresponding fixes
			for _, fix := range config.Fixes {
				err := fix.ApplyFix()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
