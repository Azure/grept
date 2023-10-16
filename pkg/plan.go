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

func (p Plan) Apply() error {
	var errs []error
	for _, fixes := range p {
		for _, fix := range fixes {
			if err := fix.ApplyFix(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors applying fixes: %v", errs)
	}
	return nil
}

type failedRule struct {
	Rule       Rule
	CheckError error
}

func (fr *failedRule) String() string {
	return fmt.Sprintf("rule.%s.%s check return failure: %s", fr.Rule.Type(), fr.Rule.Name(), fr.CheckError.Error())
}
