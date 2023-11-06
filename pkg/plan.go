package pkg

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
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
			sb.WriteString(fmt.Sprintf("fix.%s.%s would be apply:\n %s", f.Type(), f.Name(), blockToString(f)))
		}
	}
	return sb.String()
}

func (p Plan) Apply() error {
	var err error
	for _, fixes := range p {
		for _, fix := range fixes {
			if err := Eval(fix.HclSyntaxBlock(), fix); err != nil {
				err = multierror.Append(err, fmt.Errorf("rule.%s.%s(%s) eval error: %+v", fix.Type(), fix.Name(), fix.HclSyntaxBlock().Range().String(), err))
			}
		}
		if err != nil {
			return err
		}
	}

	for _, fixes := range p {
		for _, fix := range fixes {
			if applyErr := fix.ApplyFix(); applyErr != nil {
				err = multierror.Append(err, applyErr)
			}
		}
	}
	if err != nil {
		return fmt.Errorf("errors applying fixes: %+v", err)
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
