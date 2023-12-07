package pkg

import (
	"context"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"testing"
)

type copyFileFixSuite struct {
	suite.Suite
	*testBase
}

func (s *copyFileFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *copyFileFixSuite) TearDownTest() {
	s.testBase.teardown()
}

func TestCopyFileFixSuite(t *testing.T) {
	suite.Run(t, new(copyFileFixSuite))
}

func (s *copyFileFixSuite) TestCopyFileFix() {
	s.dummyFsWithFiles([]string{"/example/test"}, []string{"hello world"})
	sut := &CopyFileFix{
		BaseBlock: &BaseBlock{
			c: &Config{
				ctx: context.TODO(),
			},
			name: "test",
			id:   "test",
		},
		RuleIds: nil,
		Src:     "/example/test",
		Dest:    "/example/test2",
	}
	err := sut.ApplyFix()
	s.NoError(err)
	content, err := afero.ReadFile(s.fs, "/example/test2")
	s.NoError(err)
	s.Equal("hello world", string(content))
}