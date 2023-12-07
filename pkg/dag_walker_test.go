package pkg

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heimdalr/dag"
	"github.com/stretchr/testify/assert"
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
	t := s.T()
	content := fmt.Sprintf(`
	data "http" sample {
	    url = "%s"
    }

	rule "file_hash" sample {  
		glob = "*.txt"  
		hash = "abc123"  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_ids = [rule.file_hash.sample.id]
		paths = ["/path/to/file.txt"]  
		content = "Hello, world!"
	}  
	`, s.server.URL)

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	config, err := ParseConfig("", "", nil)
	require.NoError(t, err)
	assert.Len(t, config.dag.GetVertices(), 3)

	assertVertex(t, config.dag, "rule.file_hash.sample")
	assertVertex(t, config.dag, "data.http.sample")
	assertVertex(t, config.dag, "fix.local_file.hello_world")
}

func (s *dagSuite) TestDag_DagBlocksShouldBeConnectedWithEdgeIfThereIsReferenceBetweenTwoBlocks() {
	t := s.T()
	content := fmt.Sprintf(`
	data "http" sample {
	    url = "%s"
    }

	rule "file_hash" sample {  
		glob = "*.txt"  
		hash = sha256(data.http.sample.response_body)  
		algorithm = "sha256"  
	}  
  
	fix "local_file" hello_world{  
		rule_ids = [rule.file_hash.sample.id]
		paths = ["/path/to/file.txt"]  
		content = "Hello, world!"
	}  
	`, s.server.URL)

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	config, err := ParseConfig("", "", nil)
	require.NoError(t, err)
	assert.Equal(t, 2, config.dag.GetSize())
	roots := config.dag.GetRoots()
	assert.Len(t, roots, 1)
	assertEdge(t, config.dag, "data.http.sample", "rule.file_hash.sample")
	assertEdge(t, config.dag, "rule.file_hash.sample", "fix.local_file.hello_world")
}

func (s *dagSuite) TestDag_CycleDependencyShouldCauseError() {
	t := s.T()
	content := `
	data "http" sample {
	    url = data.http.sample2.url
    }

	data "http" sample2 {
		url = data.http.sample.url
    }
	`

	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})

	_, err := ParseConfig("", "", nil)
	require.NotNil(t, err)
	// The error message must contain both of two blocks' address so we're sure that it's about the loop.
	assert.Contains(t, err.Error(), "data.http.sample")
	assert.Contains(t, err.Error(), "data.http.sample2")
}

func assertEdge(t *testing.T, dag *dag.DAG, src, dest string) {
	from, err := dag.GetParents(dest)
	assert.NoError(t, err)
	_, ok := from[src]
	assert.True(t, ok, "cannot find edge from %s to %s", src, dest)
}

func assertVertex(t *testing.T, dag *dag.DAG, address string) {
	b, err := dag.GetVertex(address)
	assert.NoError(t, err)
	bb, ok := b.(block)
	require.True(t, ok)
	split := strings.Split(address, ".")
	name := split[len(split)-1]
	assert.Equal(t, name, bb.HclSyntaxBlock().Labels[1])
}
