package pkg

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type preConditionSuite struct {
	suite.Suite
	*testBase
}

func TestPreConditionSuite(t *testing.T) {
	suite.Run(t, new(preConditionSuite))
}

func (s *preConditionSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *preConditionSuite) TearDownTest() {
	s.teardown()
}

func (s *preConditionSuite) TestPreCondition_PassedHardcodedCondition() {
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = true
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := NewConfig("", "", nil)
	s.NoError(err)
	_, err = config.Plan()
	s.NoError(err)
	s.Len(config.RuleBlocks(), 1)
	fhr, ok := config.RuleBlocks()[0].(*FileHashRule)
	require.True(s.T(), ok)
	check, err := fhr.PreConditionCheck(new(hcl.EvalContext))
	s.NoError(err)
	s.Len(check, 0)
}

func (s *preConditionSuite) TestPreCondition_FaileddHardcodedConditionShouldFailedPlan() {
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = false
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := NewConfig("", "", nil)
	s.NoError(err)
	_, err = config.Plan()
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestPreCondition_FaileddHardcodedCondition() {
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = false
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := NewConfig("", "", nil)
	s.NoError(err)
	_, err = config.Plan()
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
	s.Len(config.RuleBlocks(), 1)
	fhr, ok := config.RuleBlocks()[0].(*FileHashRule)
	require.True(s.T(), ok)
	check, err := fhr.PreConditionCheck(new(hcl.EvalContext))
	s.NoError(err)
	s.Len(check, 1)
	s.Equal("this precondition must be true", check[0].ErrorMessage)
}
