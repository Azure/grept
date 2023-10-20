package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

func init() {
	registerRule()
	registerFix()
	registerData()
}

type block interface {
	Eval(*hclsyntax.Block) error
	Name() string
	Type() string
	Value() cty.Value
	HclSyntaxBlock() *hclsyntax.Block
}

var fixFactories = map[string]func(*Config) block{}

func registerFix() {
	fixFactories["local_file"] = func(c *Config) block {
		return &LocalFile{
			BaseFix: &BaseFix{
				c: c,
			},
		}
	}
}

var ruleFactories = map[string]func(*Config) block{}

func registerRule() {
	ruleFactories["file_hash"] = func(c *Config) block {
		return &FileHashRule{
			BaseRule: &BaseRule{
				c: c,
			},
		}
	}
}

var datasourceFactories = map[string]func(*Config) block{}

func registerData() {
	datasourceFactories["http"] = func(c *Config) block {
		return &HttpDatasource{
			BaseData: &BaseData{
				c: c,
			},
		}
	}
}

func readRequiredStringAttribute(b *hclsyntax.Block, attributeName string, ctx *hcl.EvalContext) (string, error) {
	if b == nil {
		return "", fmt.Errorf("nil Block")
	}
	a, ok := b.Body.Attributes[attributeName]
	if !ok {
		return "", fmt.Errorf("no %s in the block %s, %s", attributeName, concatLabels(b.Labels), b.Range().String())
	}
	value, diagnostics := a.Expr.Value(ctx)
	if diagnostics.HasErrors() {
		return "", fmt.Errorf("cannot evaluate expr at %s, %s", a.Expr.Range().String(), diagnostics.Error())
	}
	if value.Type() != cty.String {
		return "", fmt.Errorf("the attribute %s in the block %s (%s) is not a string", attributeName, concatLabels(b.Labels), a.Expr.Range().String())
	}
	return value.AsString(), nil
}

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

func concatLabels(labels []string) string {
	sb := strings.Builder{}
	for i, l := range labels {
		sb.WriteString(l)
		if i != len(labels)-1 {
			sb.WriteString(".")
		}
	}
	return sb.String()
}

func refresh(b block) {
	b.Eval(b.HclSyntaxBlock())
}
