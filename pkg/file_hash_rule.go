package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"hash"

	"github.com/spf13/afero"
)

var _ Rule = &FileHashRule{}

type FileHashRule struct {
	*BaseRule
	Glob      string
	Hash      string
	Algorithm string
}

func (fhr *FileHashRule) Type() string {
	return "file_hash"
}

func (fhr *FileHashRule) Value() cty.Value {
	value := fhr.BaseRule.BaseValue()
	value["glob"] = cty.StringVal(fhr.Glob)
	value["hash"] = cty.StringVal(fhr.Hash)
	value["algorithm"] = cty.StringVal(fhr.Algorithm)
	return cty.ObjectVal(value)
}

func (fhr *FileHashRule) Parse(b *hclsyntax.Block) error {
	err := fhr.BaseRule.Parse(b)
	if err != nil {
		return err
	}
	if fhr.Glob, err = readRequiredStringAttribute(b, "glob", fhr.EvalContext()); err != nil {
		return err
	}
	if fhr.Hash, err = readRequiredStringAttribute(b, "hash", fhr.EvalContext()); err != nil {
		return err
	}
	if fhr.Algorithm, err = readOptionalStringAttribute(b, "algorithm", fhr.EvalContext()); err != nil {
		return err
	}
	if fhr.Algorithm == "" {
		fhr.Algorithm = "sha1"
	}
	switch fhr.Algorithm {
	case "md5", "sha1", "sha256", "sha512":
		// valid
	default:
		return fmt.Errorf("invalid algorithm: %s", fhr.Algorithm)
	}
	return nil
}

func (fhr *FileHashRule) Check() (error, error) {
	// Use Glob to find files matching the path pattern
	fs := FsFactory()
	files, err := afero.Glob(fs, fhr.Glob)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return fmt.Errorf("no files match path pattern: %s", fhr.Glob), nil
	}

	for _, file := range files {
		fileData, err := afero.ReadFile(fs, file)
		if err != nil {
			return nil, err
		}

		// Calculate the hash of the file data
		var h hash.Hash
		switch fhr.Algorithm {
		case "md5":
			h = md5.New()
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		case "sha1":
			fallthrough
		default: // Default to sha1
			h = sha1.New()
		}
		h.Write(fileData)
		computedHash := fmt.Sprintf("%x", h.Sum(nil))

		if computedHash == fhr.Hash {
			return nil, nil
		}
	}

	return fmt.Errorf("no file with glob %s and hash %s found", fhr.Glob, fhr.Hash), nil
}

func (fhr *FileHashRule) Validate() error {
	if fhr.Glob == "" {
		return fmt.Errorf("glob is required")
	}

	if fhr.Hash == "" {
		return fmt.Errorf("hash is required")
	}

	if fhr.Algorithm != "" {
		switch fhr.Algorithm {
		case "md5", "sha1", "sha256", "sha512":
			// valid
		default:
			return fmt.Errorf("invalid algorithm: %s", fhr.Algorithm)
		}
	}

	return nil
}
