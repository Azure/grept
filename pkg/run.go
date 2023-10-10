package pkg

import (
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

	//ruleMap := make(map[string]Rule)
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
		Functions: map[string]function.Function{},
	}
	// First loop: parse all rule blocks
	for _, block := range body.Blocks {
		if block.Type == "rule" {
			t := block.Labels[0]
			rule := RuleFactories[t](ctx)
			err := rule.Parse(block)
			if err != nil {
				return nil, err
			}
			config.Rules = append(config.Rules, rule)
			//ruleMap[fmt.Sprintf("%s.%s", t, n)] = rule
		}
	}

	// Second loop: parse all fix blocks
	for _, block := range body.Blocks {
		if block.Type == "fix" {
			t := block.Labels[0]
			fix := FixFactories[t](ctx)
			err := fix.Parse(block)
			if err != nil {
				return nil, err
			}
			//if _, exists := ruleMap[fix.GetRule()]; !exists {
			//	return nil, fmt.Errorf("rule %s does not exist", fix.GetRule())
			//}
			config.Fixes = append(config.Fixes, fix)
		}
	}

	return &config, nil
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
