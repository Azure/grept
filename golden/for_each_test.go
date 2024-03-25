package golden

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type forEachTestSuite struct {
	suite.Suite
	*testBase
}

func TestForEachTestSuite(t *testing.T) {
	suite.Run(t, new(forEachTestSuite))
}

func (s *forEachTestSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *forEachTestSuite) SetupSubTest() {
	s.SetupTest()
}

func (s *forEachTestSuite) TearDownTest() {
	s.teardown()
}

func (s *forEachTestSuite) TearDownSubTest() {
	s.TearDownTest()
}

func (s *forEachTestSuite) TestForEachBlockWithAttributeThatHasDefaultValue() {
	config := `	
	data "dummy" "sample" {
		for_each = toset([1,2,3])
	}
`
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{config})
	c, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	_, err = RunDummyPlan(c)
	s.NoError(err)
	for _, b := range blocks(c) {
		data := b.(*DummyData)
		s.Equal("default_value", data.AttributeWithDefaultValue)
	}
}

func (s *forEachTestSuite) TestLocals_locals_as_for_each() {
	code := `
locals {
  numbers = toset([1,2,3])
}

data "dummy" foo {
	for_each = local.numbers
}
`
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{code})
	c, err := BuildDummyConfig("/", "", nil)
	s.NoError(err)
	p, err := RunDummyPlan(c)
	s.NoError(err)
	s.Len(p.Datas, 3)
}

func (s *forEachTestSuite) TestLocals_data_output_as_foreach() {
	code := `
data "dummy" foo {
	data = {
		"1" = "one"
		"2" = "two"
		"3" = "three"
	}
}

resource "dummy" bar {
	for_each = data.dummy.foo.data
}
`
	s.dummyFsWithFiles([]string{"test.hcl"}, []string{code})
	c, err := BuildDummyConfig("/", "", nil)
	s.NoError(err)
	p, err := RunDummyPlan(c)
	s.NoError(err)
	s.Len(p.Resources, 3)
}
