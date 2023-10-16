package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestPlanFunc_NoCheckFailure(t *testing.T) {
	expectedContent := "Mock server response"
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, expectedContent)
	}))
	defer ts.Close()

	// Mock config
	configContent := `  
		data "http" "test" {  
			url = "` + ts.URL + `"  
		}  
  
		rule "file_hash" "test" {  
			glob = "test.txt"  
			hash = sha1(data.http.test.response_body)  
		}  
	`

	mockFs := afero.NewMemMapFs()
	stub := gostub.Stub(&pkg.FsFactory, func() afero.Fs {
		return mockFs
	})
	defer stub.Reset()

	_ = afero.WriteFile(mockFs, "test.txt", []byte(expectedContent), 0644)
	_ = afero.WriteFile(mockFs, "test_config.hcl", []byte(configContent), 0644)

	// Redirect Stdout
	r, w, _ := os.Pipe()
	stub.Stub(&os.Stdout, w)

	cmd := NewPlanCmd(context.TODO())
	// Run function
	cmd.Run(nil, []string{"test_config.hcl"})

	// Reset Stdout
	w.Close()

	// Read Stdout
	out, _ := io.ReadAll(r)
	output := string(out)

	assert.Contains(t, output, "All rule checks successful, nothing to do.")
}

func TestPlanFunc_CheckFailure(t *testing.T) {
	expectedContent := "Mock server response"
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, expectedContent)
	}))
	defer ts.Close()

	// Mock config
	configContent := `    
		data "http" "test" {    
			url = "` + ts.URL + `"    
		}    
    
		rule "file_hash" "test" {    
			glob = "test.txt"    
			hash = sha1(data.http.test.response_body)    
		}  
  
		fix "local_file" "test" {  
			path = "test.txt"  
			content = data.http.test.response_body  
			rule_id = rule.file_hash.test.id  
		}  
	`

	mockFs := afero.NewMemMapFs()
	stub := gostub.Stub(&pkg.FsFactory, func() afero.Fs {
		return mockFs
	})
	defer stub.Reset()

	_ = afero.WriteFile(mockFs, "test.txt", []byte("incorrect content"), 0644)
	_ = afero.WriteFile(mockFs, "test_config.hcl", []byte(configContent), 0644)

	// Redirect Stdout
	r, w, _ := os.Pipe()
	stub.Stub(&os.Stdout, w)

	cmd := NewPlanCmd(context.TODO())
	// Run function
	cmd.Run(nil, []string{"test_config.hcl"})

	// Reset Stdout
	w.Close()

	// Read Stdout
	out, _ := io.ReadAll(r)
	output := string(out)

	assert.Contains(t, output, "rule.file_hash.test check return failure:")
	assert.Contains(t, output, "fix.local_file.test would be apply:")
	assert.Contains(t, output, `"Content":"Mock server response"`)
}
