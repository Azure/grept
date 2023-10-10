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
	fs        afero.Fs
	Glob      string
	Hash      string
	Algorithm string
}

func (fhr *FileHashRule) Parse(b *hclsyntax.Block) error {
	err := fhr.BaseRule.Parse(b)
	if err != nil {
		return err
	}
	fhr.Glob, err = readRequiredStringAttribute(b, "glob", fhr.ctx)
	if err != nil {
		return err
	}
	fhr.Hash, err = readRequiredStringAttribute(b, "hash", fhr.ctx)
	if err != nil {
		return err
	}
	fhr.Algorithm, _ = readOptionalStringAttribute(b, "algorithm", fhr.ctx)
	if fhr.Algorithm == "" {
		fhr.Algorithm = "sha1"
	}
	switch fhr.Algorithm {
	case "md5", "sha1", "sha256", "sha512":
		// valid
	default:
		return fmt.Errorf("invalid algorithm: %s", fhr.Algorithm)
	}

	blockAddress := concatLabels(b.Labels)
	fhr.ctx.Variables[blockAddress] = cty.StringVal(blockAddress)
	m, ok := fhr.ctx.Variables[b.Labels[0]]
	if !ok {
		m = cty.ObjectVal(map[string]cty.Value{
			b.Labels[1]: cty.StringVal(blockAddress),
		})
		fhr.ctx.Variables[b.Labels[0]] = m
		return nil
	}
	m.AsValueMap()[b.Labels[1]] = cty.StringVal(blockAddress)
	return nil
}

func (fhr *FileHashRule) Check() error {
	// Use Glob to find files matching the path pattern
	files, err := afero.Glob(fhr.fs, fhr.Glob)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no files match path pattern: %s", fhr.Glob)
	}

	for _, file := range files {
		fileData, err := afero.ReadFile(fhr.fs, file)
		if err != nil {
			return err
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
			return nil
		}
	}

	return fmt.Errorf("no file with glob %s and hash %s found", fhr.Glob, fhr.Hash)
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
