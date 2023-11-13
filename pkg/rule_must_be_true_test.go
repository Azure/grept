package pkg

import (
	"context"
	"fmt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zclconf/go-cty/cty"
	"os"
	"path/filepath"
	"testing"
)

type mustBeTrueRuleSuite struct {
	suite.Suite
	*testBase
}

func TestMustBeTrueRuleSuite(t *testing.T) {
	suite.Run(t, new(mustBeTrueRuleSuite))
}

func (s *mustBeTrueRuleSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *mustBeTrueRuleSuite) TearDownTest() {
	s.teardown()
}

func (s *mustBeTrueRuleSuite) TestMustBeTrueRule_Check() {
	t := s.T()
	tests := []struct {
		name        string
		rule        *MustBeTrueRule
		expectError bool
	}{
		{
			name: "Condition is true",
			rule: &MustBeTrueRule{
				BaseBlock:    &BaseBlock{},
				Condition:    true,
				ErrorMessage: "",
			},
			expectError: false,
		},
		{
			name: "Condition is false",
			rule: &MustBeTrueRule{
				BaseBlock:    &BaseBlock{},
				Condition:    false,
				ErrorMessage: "Test error message",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, runTimeErr := tt.rule.Check()
			require.NoError(t, runTimeErr)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "assertion failed")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func (s *mustBeTrueRuleSuite) TestMustBeTrueRule_Eval() {
	fs := s.fs
	t := s.T()
	expectedContent := "hello world"
	temp, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	_, err = temp.WriteString(expectedContent)
	require.NoError(t, err)
	err = temp.Close()
	defer func() {
		s := temp.Name()
		_ = os.Remove(s)
	}()

	_ = afero.WriteFile(fs, "/test.grept.hcl", []byte(fmt.Sprintf(`
	rule "must_be_true" sample {
		condition = file("%s") == "hello world"
	}
`, filepath.ToSlash(temp.Name()))), 0644)

	c, err := ParseConfig("/", "/", context.TODO())
	require.NoError(t, err)
	plan, err := c.Plan()
	require.NoError(t, err)
	assert.Empty(t, plan.FailedRules)
}

func (s *mustBeTrueRuleSuite) TestMustBeTrueRule_Value() {
	t := s.T()
	mustBeTrueRule := &MustBeTrueRule{
		BaseBlock:    &BaseBlock{},
		Condition:    true,
		ErrorMessage: "Test error message",
	}

	value := mustBeTrueRule.Values()

	assert.Equal(t, map[string]cty.Value{
		"condition":     cty.BoolVal(mustBeTrueRule.Condition),
		"error_message": cty.StringVal(mustBeTrueRule.ErrorMessage),
	}, value)
}
