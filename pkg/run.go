package pkg

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type Config struct {
	Rules []Rule
	Fixes []Fix
}

func ParseConfig(fn, content string) (*Config, error) {
	var config Config

	file, diag := hclsyntax.ParseConfig([]byte(content), fn, hcl.InitialPos)
	if diag.HasErrors() {
		return nil, diag
	}
	body := file.Body.(*hclsyntax.Body)
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}
	var err error
	// First loop: parse all rule blocks
	for _, block := range body.Blocks {
		if block.Type != "rule" && block.Type != "fix" {
			err = multierror.Append(err, fmt.Errorf("invalid block type: %s %s", block.Type, block.Range().String()))
		}
		if block.Type == "rule" {
			t := block.Labels[0]
			rf, ok := RuleFactories[t]
			if !ok {
				err = multierror.Append(err, fmt.Errorf("unregistered rule: %s, %s", t, block.Range().String()))
				continue
			}
			rule := rf(ctx)
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
			t := block.Labels[0]
			ff, ok := FixFactories[t]
			if !ok {
				err = multierror.Append(err, fmt.Errorf("unregistered fix: %s, %s", t, block.Range().String()))
				continue
			}
			fix := ff(ctx)
			parseError := fix.Parse(block)
			if parseError != nil {
				err = multierror.Append(err, parseError)
				continue
			}
			config.Fixes = append(config.Fixes, fix)
		}
	}

	return &config, err
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
