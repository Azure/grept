package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"strings"
)

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
