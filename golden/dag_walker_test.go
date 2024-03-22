package golden

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type dagSuite struct {
	suite.Suite
	*testBase
	server *httptest.Server
}

func TestDagSuite(t *testing.T) {
	suite.Run(t, new(dagSuite))
}

func (s *dagSuite) SetupTest() {
	s.testBase = newTestBase()
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Expected content"))
	}))
}

func (s *dagSuite) TearDownTest() {
	s.teardown()
	s.server.Close()
}

func (s *dagSuite) TestDag_DagVertex() {
	content := `
	data "dummy" foo {}

	resource "dummy" bar {}  
	`

	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})

	config, err := BuildDummyConfig("", "", nil)
	s.NoError(err)
	d := newDag()
	err = d.buildDag(blocks(config))
	s.NoError(err)
	s.Len(d.GetVertices(), 2)

	assertVertex(s.T(), d, "data.dummy.foo")
	assertVertex(s.T(), d, "resource.dummy.bar")
}

func (s *dagSuite) TestDag_DagBlocksShouldBeConnectedWithEdgeIfThereIsReferenceBetweenTwoBlocks() {
	t := s.T()
	content := `
	data "dummy" foo {
	    data = {
			key = "value"
		}
    }

	resource "dummy" foo {
		tags = data.dummy.foo.data
	}  
	resource "dummy" bar {
		tags = merge(data.dummy.foo.data, resource.dummy.foo.tags)
	}
	`

	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})

	config, err := BuildDummyConfig("", "", nil)
	require.NoError(t, err)
	dag := newDag()
	err = dag.buildDag(blocks(config))
	require.NoError(t, err)
	assert.Equal(t, 3, dag.GetSize())
	roots := dag.GetRoots()
	assert.Len(t, roots, 1)
	assertEdge(t, dag, "data.dummy.foo", "resource.dummy.foo")
	assertEdge(t, dag, "data.dummy.foo", "resource.dummy.foo")
	assertEdge(t, dag, "resource.dummy.foo", "resource.dummy.bar")
}

func (s *dagSuite) TestDag_CycleDependencyShouldCauseError() {
	content := `
	data "dummy" sample {
	    data = data.dummy.sample2.data
    }

	data "dummy" sample2 {
		data = data.dummy.sample.data
    }
	`

	s.dummyFsWithFiles([]string{"test.hcl"}, []string{content})

	_, err := BuildDummyConfig("", "", nil)
	s.NotNil(err)
	// The error message must contain both of two blocks' address so we're sure that it's about the loop.
	s.Contains(err.Error(), "data.dummy.sample")
	s.Contains(err.Error(), "data.dummy.sample2")
}

func assertEdge(t *testing.T, dag *Dag, src, dest string) {
	from, err := dag.GetParents(dest)
	assert.NoError(t, err)
	_, ok := from[src]
	assert.True(t, ok, "cannot find edge from %s to %s", src, dest)
}

func assertVertex(t *testing.T, dag *Dag, address string) {
	b, err := dag.GetVertex(address)
	assert.NoError(t, err)
	bb, ok := b.(Block)
	require.True(t, ok)
	split := strings.Split(address, ".")
	name := split[len(split)-1]
	assert.Equal(t, name, bb.HclBlock().Labels[1])
}
