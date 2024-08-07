package pkg

import (
	"fmt"
	"github.com/Azure/golden"
	"github.com/ahmetb/go-linq/v3"
	"strings"
	"sync"
)

func RunGreptPlan(c *GreptConfig) (*GreptPlan, error) {
	err := c.RunPlan()
	if err != nil {
		return nil, err
	}

	plan := newPlan(c)
	for _, rb := range golden.Blocks[Rule](c) {
		checkErr := rb.CheckError()
		if checkErr == nil {
			continue
		}
		plan.addRule(&FailedRule{
			Rule:       rb,
			CheckError: checkErr,
		})
		for _, fb := range golden.Blocks[Fix](c) {
			if linq.From(fb.GetRuleIds()).Contains(rb.Id()) {
				plan.addFix(fb)
			}
		}
	}

	return plan, nil
}

var _ golden.Plan = &GreptPlan{}

type GreptPlan struct {
	FailedRules []*FailedRule
	Fixes       map[string]Fix
	c           *GreptConfig
	mu          sync.Mutex
}

func newPlan(c *GreptConfig) *GreptPlan {
	return &GreptPlan{
		c:     c,
		Fixes: make(map[string]Fix),
	}
}

func (p *GreptPlan) String() string {
	sb := strings.Builder{}
	for _, r := range p.FailedRules {
		sb.WriteString(r.String())
		sb.WriteString("\n---\n")
	}
	for _, f := range p.Fixes {
		sb.WriteString(fmt.Sprintf("%s would be apply:\n %s\n", f.Address(), golden.BlockToString(f)))
		sb.WriteString("\n---\n")
	}

	return sb.String()
}

func (p *GreptPlan) Apply() error {
	addresses := make(map[string]struct{})
	for _, fix := range p.Fixes {
		addresses[fix.Address()] = struct{}{}
	}
	if err := golden.Traverse[Fix](p.c.BaseConfig, func(fix Fix) error {
		if decodeErr := golden.Decode(fix); decodeErr != nil {
			return fmt.Errorf("rule.%s.%s(%s) decode error: %+v", fix.Type(), fix.Name(), fix.HclBlock().Range().String(), decodeErr)
		}
		return nil
	}); err != nil {
		return err
	}
	if err := golden.Traverse[Fix](p.c.BaseConfig, func(fix Fix) error {
		return fix.Apply()
	}); err != nil {
		return err
	}
	return nil
}

func (p *GreptPlan) addRule(fr *FailedRule) {
	p.mu.Lock()
	p.FailedRules = append(p.FailedRules, fr)
	p.mu.Unlock()
}

func (p *GreptPlan) addFix(f Fix) {
	p.mu.Lock()
	p.Fixes[f.Id()] = f
	p.mu.Unlock()
}

type FailedRule struct {
	Rule
	CheckError error
}

func (fr *FailedRule) String() string {
	address := fr.Address()
	return fmt.Sprintf("%s check return failure: %s", address, fr.CheckError.Error())
}
