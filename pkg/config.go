package pkg

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Rules []Rule
type Fixes []Fix

func (rs Rules) Values() cty.Value {
	if len(rs) == 0 {
		return cty.EmptyObjectVal
	}
	res := map[string]cty.Value{}
	valuesMap := map[string]map[string]cty.Value{}

	for _, r := range rs {
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
	Rules Rules
	Fixes Fixes
}

func (c *Config) EvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"rule": c.Rules.Values(),
		},
	}
}

func ParseConfig(fn, content string) (*Config, error) {
	config := &Config{}

	file, diag := hclsyntax.ParseConfig([]byte(content), fn, hcl.InitialPos)
	if diag.HasErrors() {
		return nil, diag
	}
	body := file.Body.(*hclsyntax.Body)
	//ctx := &hcl.EvalContext{
	//	Variables: map[string]cty.Value{},
	//	Functions: map[string]function.Function{},
	//}
	var err error
	// First loop: parse all rule blocks
	for _, block := range body.Blocks {
		if block.Type != "rule" && block.Type != "fix" {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", block.Type, block.Range().String()))
		}
		if block.Type == "rule" {
			if len(block.Labels) != 2 {
				err = multierror.Append(err, fmt.Errorf("invalid labels for rule %s, expect labels with length 2 (%s)", concatLabels(block.Labels), block.Range().String()))
				continue
			}
			t := block.Labels[0]
			rf, ok := RuleFactories[t]
			if !ok {
				err = multierror.Append(err, fmt.Errorf("unregistered rule: %s, %s", t, block.Range().String()))
				continue
			}
			rule := rf(config.EvalContext())
			parseError := rule.Parse(block)
			if parseError != nil {
				err = multierror.Append(err, parseError)
				continue
			}
			config.Rules = append(config.Rules, rule)
		}
	}

	// Second loop: parse all fix blocks
	for _, block := range body.Blocks {
		if block.Type == "fix" {
			if len(block.Labels) != 2 {
				err = multierror.Append(err, fmt.Errorf("invalid labels for fix %s, expect labels with length 2 (%s)", concatLabels(block.Labels), block.Range().String()))
				continue
			}
			t := block.Labels[0]
			ff, ok := FixFactories[t]
			if !ok {
				err = multierror.Append(err, fmt.Errorf("unregistered fix: %s, %s", t, block.Range().String()))
				continue
			}
			fix := ff(config.EvalContext())
			parseError := fix.Parse(block)
			if parseError != nil {
				err = multierror.Append(err, parseError)
				continue
			}
			config.Fixes = append(config.Fixes, fix)
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
