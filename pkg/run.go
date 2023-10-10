package pkg

import (
	"github.com/Azure/grept/pkg/fixes"
	"github.com/Azure/grept/pkg/rules"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Config struct {
	Rules []rules.Rule
	Fixes []fixes.Fix
}

func ParseConfig(fn, content string) (*Config, error) {
	var config Config

	file, diag := hclsyntax.ParseConfig([]byte(content), fn, hcl.InitialPos)
	if diag.HasErrors() {
		return nil, diag
	}

	body := file.Body.(*hclsyntax.Body)

	for _, block := range body.Blocks {
		switch block.Type {
		case "rule":
			t := block.Labels[0]
			rule := rules.RuleFactories[t]()
			diag := gohcl.DecodeBody(block.Body, nil, rule)
			if diag.HasErrors() {
				return nil, diag
			}
			config.Rules = append(config.Rules, rule)
		case "fix":
			t := block.Labels[0]
			fix := fixes.FixFactories[t]()
			diag := gohcl.DecodeBody(block.Body, nil, fix)
			if diag.HasErrors() {
				return nil, diag
			}
			config.Fixes = append(config.Fixes, fix)
		default:
			// handle unknown block
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
