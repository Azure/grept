package pkg

import (
	"fmt"
	"strings"
)

type Plan map[*failedRule]Fixes

func (p Plan) String() string {
	sb := strings.Builder{}
	for fr, fixes := range p {
		sb.WriteString(fr.String())
		sb.WriteString("\n")
		for _, f := range fixes {
			sb.WriteString("  ")
			sb.WriteString(fmt.Sprintf("fix.%s.%s would be apply:\n %s", f.Type(), f.Name(), FixToString(f)))
		}
	}
	return sb.String()
}

type failedRule struct {
	Rule       Rule
	CheckError error
}

func (fr *failedRule) String() string {
	return fmt.Sprintf("rule.%s.%s check return failure: %s", fr.Rule.Type(), fr.Rule.Name(), fr.CheckError.Error())
}
