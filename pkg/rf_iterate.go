package pkg

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type refIterator func(t []hcl.Traverser, i int) *string

var refIters = map[string]refIterator{
	"data":  dataIterator,
	"rule":  ruleIterator,
	"fix":   fixIterator,
	"local": localIterator,
}

var localIterator = iterator("local", 2)
var dataIterator = iterator("data", 3)
var ruleIterator = iterator("rule", 3)
var fixIterator = iterator("fix", 3)

func iterator(keyword string, addressLength int) refIterator {
	return func(ts []hcl.Traverser, i int) *string {
		if len(ts) == 0 {
			return nil
		}
		if name(ts[i]) != keyword {
			return nil
		}
		if len(ts) < i+addressLength {
			return nil
		}
		remain := addressLength
		sb := strings.Builder{}
		for j := i; remain > 0; j++ {
			sb.WriteString(name(ts[j]))
			remain--
			if remain > 0 {
				sb.WriteString(".")
			}
		}
		r := sb.String()
		return &r
	}
}

func name(t hcl.Traverser) string {
	switch tp := t.(type) {
	case hcl.TraverseRoot:
		{
			return tp.Name
		}
	case hcl.TraverseAttr:
		{
			return tp.Name
		}
	default:
		{
			return ""
		}
	}
}
