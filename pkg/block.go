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
	Parse(*hclsyntax.Block) error
}

var fixFactories = map[string]func(*hcl.EvalContext) block{}

func registerFix() {
	fixFactories["local_file"] = func(ctx *hcl.EvalContext) block {
		return &LocalFile{
			BaseFix: &BaseFix{
				ctx: ctx,
			},
		}
	}
}

var ruleFactories = map[string]func(*hcl.EvalContext) block{}

func registerRule() {
	ruleFactories["file_hash"] = func(ctx *hcl.EvalContext) block {
		return &FileHashRule{
			BaseRule: &BaseRule{
				ctx: ctx,
			},
		}
	}
}

var datasourceFactories = map[string]func(ctx *hcl.EvalContext) block{}

func registerData() {
	datasourceFactories["http"] = func(ctx *hcl.EvalContext) block {
		return &HttpDatasource{
			BaseData: &BaseData{
				ctx: ctx,
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

func readOptionalStringAttribute(b *hclsyntax.Block, attributeName string, ctx *hcl.EvalContext) (string, error) {
	if b == nil {
		return "", fmt.Errorf("nil Block")
	}
	a, ok := b.Body.Attributes[attributeName]
	if !ok {
		return "", nil
	}
	value, diagnostics := a.Expr.Value(ctx)
	if diagnostics.HasErrors() {
		return "", diagnostics
	}
	if value.Type() != cty.String {
		return "", fmt.Errorf("the attribute %s in the block %s is not a string", attributeName, concatLabels(b.Labels))
	}
	return value.AsString(), nil
}

func readOptionalMapAttribute(b *hclsyntax.Block, attributeName string, ctx *hcl.EvalContext) (map[string]string, error) {
	if b == nil {
		return nil, fmt.Errorf("nil Block")
	}
	a, ok := b.Body.Attributes[attributeName]
	if !ok {
		return nil, nil
	}
	value, diagnostics := a.Expr.Value(ctx)
	if diagnostics.HasErrors() {
		return nil, diagnostics
	}
	if value.Type() != cty.Map(cty.String) && !objectIsMapOfString(value.Type()) {
		return nil, fmt.Errorf("the attribute %s in the block %s is not a map of string", attributeName, concatLabels(b.Labels))
	}
	r := make(map[string]string)
	for k, v := range value.AsValueMap() {
		r[k] = v.AsString()
	}
	return r, nil
}

func objectIsMapOfString(t cty.Type) bool {
	if !t.IsObjectType() {
		return false
	}
	for _, at := range t.AttributeTypes() {
		if at != cty.String {
			return false
		}
	}
	return true
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
