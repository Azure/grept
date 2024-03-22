package golden

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
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
    data "dummy" foo {
        precondition {
            condition = true
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(config)
	s.NoError(err)
	ds := Blocks[TestData](config)
	s.Len(ds, 1)
	d, ok := ds[0].(*DummyData)
	require.True(s.T(), ok)
	check, err := d.PreConditionCheck(new(hcl.EvalContext))
	s.NoError(err)
	s.Len(check, 0)
}

func (s *preConditionSuite) TestPreCondition_FailedHardcodedConditionShouldFailedPlan() {
	content := `
	data "dummy" foo {
        precondition {
            condition = false
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestPreCondition_FailedHardcodedCondition() {
	content := `
    data "dummy" foo {
        precondition {
            condition = false
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
	ds := Blocks[TestData](config)
	s.Len(ds, 1)
	d, ok := ds[0].(*DummyData)
	require.True(s.T(), ok)
	check, err := d.PreConditionCheck(new(hcl.EvalContext))
	s.NoError(err)
	s.Len(check, 1)
	s.Equal("this precondition must be true", check[0].ErrorMessage)
}

func (s *preConditionSuite) TestPreCondition_FunctionCallInCondition() {
	s.T().Setenv("KEY", "VALUE")
	content := `
	data "dummy" foo {
        precondition {
            condition = env("KEY") != "VALUE"
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestMultiplePreConditions_containFailedCheck() {
	// Test case when one of the preconditions fails
	content := `
        data "dummy" foo {
            precondition {
                condition = true
            }
            precondition {
                condition = false
                error_message = "this precondition must be true"
            }
        }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestMultiplePreConditions_allPassedCheck() {
	// Test case when all preconditions pass
	content := `
        data "dummy" foo {
            precondition {
                condition = true
            }
            precondition {
                condition = true
            }
        }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(config)
	s.NoError(err)
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttribute() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	sampleConfig := `
    data "dummy" "foo" {
        data = {
			key = "value"
		}
    }

    data "dummy" "foo1" {
        precondition {
            condition = data.dummy.foo.data["key"] == "value"
            error_message = "Precondition check failed"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunDummyPlan(config)
	s.NoError(err) // Expect no error as the precondition should pass
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttributeFailedCheck() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	sampleConfig := `
    data "dummy" "foo" {
        data = {
			key = "value"
		}
    }

    data "dummy" "foo1" {
        precondition {
            condition = data.dummy.foo.data["key"] != "value"
            error_message = "Precondition check failed"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunDummyPlan(config)
	s.Contains(err.Error(), "Precondition check failed")
}

func (s *preConditionSuite) TestPreCondition_ReferMultipleBlockAttributes() {
	// Define a sample config for testing
	sampleConfig := `
    data "dummy" "foo1" {
        data = {
			key = "value"
		}
    }

    data "dummy" "foo2" {
        data = {
			key = "value"
		}
    }

    data "dummy" "bar" {
        precondition {
            condition = data.dummy.foo1.data["key"] == data.dummy.foo2.data["key"]
            error_message = "Precondition check failed"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunDummyPlan(config)
	s.NoError(err) // Expect no error as the precondition should pass
}

func (s *preConditionSuite) TestPreCondition_ReferMultipleBlockAttributesFailedCheck() {

	// Define a sample config for testing
	sampleConfig := `
    data "dummy" "foo1" {
        data = {
			key = "value"
		}
    }

    data "dummy" "foo2" {
        data = {
			key = "value"
		}
    }

    data "dummy" "bar" {
        precondition {
            condition = data.dummy.foo1.data["key"] != data.dummy.foo2.data["key"]
            error_message = "Precondition check failed"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunDummyPlan(config)
	s.Contains(err.Error(), "Precondition check failed")
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttributeWithForEach() {
	// Define a sample config for testing
	sampleConfig := `
        locals {
            items = toset(["item1", "item2", "item3"])
        }

        data "dummy" "foo" {
            for_each = local.items
			data = {
				key = each.value
			}
        }

        data "dummy" bar {
            precondition {
                condition = data.dummy.foo["item1"].data["key"] == "item1"
                error_message = "Precondition check failed"
            }
        }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunDummyPlan(config)
	s.NoError(err) // Expect no error as the precondition should pass
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttributeWithForEach_FailedCheck() {
	// Define a sample config for testing
	sampleConfig := `
        locals {
            items = toset(["item1", "item2", "item3"])
        }

        data "dummy" "foo" {
            for_each = local.items
			data = {
				key = each.value
			}
        }

        data "dummy" bar {
            precondition {
                condition = data.dummy.foo["item2"].data["key"] == "item1"
                error_message = "Precondition check failed"
            }
        }
    `
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunDummyPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "Precondition check failed")
}
