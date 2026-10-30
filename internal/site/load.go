package site

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and strictly decodes the content file at path.
func Load(path string) (*Site, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// Parse strictly decodes site YAML: unknown fields are errors, so a typo
// fails the build instead of silently dropping content.
func Parse(r io.Reader) (*Site, error) {
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	var s Site
	if err := dec.Decode(&s); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("content is empty")
		}
		return nil, err
	}
	return &s, nil
}
