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
