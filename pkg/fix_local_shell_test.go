package pkg

import (
	"context"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/google/uuid"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

type localExecFixSuite struct {
	suite.Suite
	*testBase
}

func TestLocalExecFixSuite(t *testing.T) {
	suite.Run(t, new(localExecFixSuite))
}

func (s *localExecFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *localExecFixSuite) TearDownTest() {
	s.teardown()
}

func (s *localExecFixSuite) TestLocalExecShell_CommandValidate() {
	cases := []struct {
		desc      string
		f         *LocalShellFix
		wantError bool
	}{
		{
			desc: "Inlines only",
			f: &LocalShellFix{
				Inlines: []string{"echo hello"},
			},
			wantError: false,
		},
		{
			desc: "Script only",
			f: &LocalShellFix{
				Inlines: nil,
				Script:  "./bash.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript only",
			f: &LocalShellFix{
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript http only",
			f: &LocalShellFix{
				RemoteScript: "http://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "non-http RemoteScript",
			f: &LocalShellFix{
				RemoteScript: "s3::https://s3.amazonaws.com/bucket/foo",
			},
			wantError: true,
		},
		{
			desc: "Inlines&Script",
			f: &LocalShellFix{
				Inlines: []string{"echo hello"},
				Script:  "./bash.sh",
			},
			wantError: true,
		},
		{
			desc: "Script&RemoteScrip",
			f: &LocalShellFix{
				Inlines:      nil,
				Script:       "./bash.sh",
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc: "Inlines&RemoteScript",
			f: &LocalShellFix{
				Inlines:      []string{"echo hello"},
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc:      "No command",
			f:         &LocalShellFix{},
			wantError: true,
		},
		{
			desc: "InlineShebang Only",
			f: &LocalShellFix{
				Script:        "/test.sh",
				InlineShebang: "#!/bin/sh",
			},
			wantError: true,
		},
		{
			desc: "InlineShebang with inlines",
			f: &LocalShellFix{
				InlineShebang: "#!/bin/sh",
				Inlines: []string{
					"echo hello",
				},
			},
			wantError: false,
		},
		{
			desc: "Invalid only_on",
			f: &LocalShellFix{
				Inlines: []string{
					"echo hello",
				},
				OnlyOn: []string{
					"Linux", // should be linux
				},
			},
			wantError: true,
		},
		{
			desc: "Valid only_on",
			f: &LocalShellFix{
				Inlines: []string{
					"echo hello",
				},
				OnlyOn: []string{
					"linux",
					"windows",
					"darwin",
				},
			},
			wantError: false,
		},
	}
	for _, c := range cases {
		s.Run(c.desc, func() {
			err := validate.Struct(*c.f)
			if c.wantError {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})
	}
}

func (s *localExecFixSuite) TestLocalExecShell_ShouldReturnDirectlyIfOnlyOnMismatch() {
	quit := false
	stub := gostub.Stub(&stopByOnlyOnStub, func() {
		quit = true
	})
	defer stub.Reset()

	onlyOn := []string{"linux", "windows", "darwin"}
	linq.From(onlyOn).Where(func(i interface{}) bool {
		return i.(string) != runtime.GOOS // exclude the current os
	}).ToSlice(&onlyOn)

	sut := &LocalShellFix{
		Inlines: []string{"echo hello"},
		OnlyOn:  onlyOn,
	}
	err := sut.ApplyFix()
	assert.NoError(s.T(), err)
	assert.True(s.T(), quit)
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_Inlines() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &Config{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		Inlines:        []string{`echo "Hello, World!"`},
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err := fix.ApplyFix()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// Check that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_cript() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	tmpScript, err := os.CreateTemp("", "grep-test")
	_ = os.Chmod(tmpScript.Name(), 0700)
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tmpScript.Name())
	}()
	_, _ = tmpScript.WriteString(`#!/bin/sh
		echo "Hello, World!"
`)
	_ = tmpScript.Close()

	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &Config{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		Script:         tmpScript.Name(),
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err = fix.ApplyFix()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// Check that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_RemoteScript() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintln(w, `#!/bin/sh
		echo "Hello, World!"`)
	}))
	defer ts.Close()

	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &Config{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		RemoteScript:   fmt.Sprintf("%s/test.sh", ts.URL),
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err := fix.ApplyFix()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// Check that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	rand := uuid.NewString()
	t.Setenv("TMP_VAR", rand)
	temp, err := os.CreateTemp("", "test_grept")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(temp.Name())
	}()
	err = temp.Close()
	require.NoError(t, err)
	hcl := fmt.Sprintf(`
	rule "must_be_true" "example" {
		condition = false
	}

	fix "local_shell" "example" {
		rule_id = rule.must_be_true.example.id
		inlines = [
			"echo ${env("TMP_VAR")}>%s",
		]
	}
`, temp.Name())
	s.dummyFsWithFiles([]string{"/example/test.grept.hcl"}, []string{hcl})
	config, err := ParseConfig("/example", context.TODO())
	require.NoError(t, err)
	plan, err := config.Plan()
	require.NoError(t, err)
	err = plan.Apply()
	require.NoError(t, err)
	tmpFileForRead, err := os.ReadFile(temp.Name())
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSuffix(string(tmpFileForRead), "\n"), rand)
}
