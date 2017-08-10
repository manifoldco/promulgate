package artifact

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Release represents a release object
type Release struct {
	Tag  string // tag
	Body string // markdown formatted release notes
}

// File represents a file on disk
type File struct {
	Name string
	Path string
	Type string

	Data io.Reader
}

// FindZips finds the pre-built zips for the given tag in the given directory
func FindZips(path, project, tag string) ([]File, error) {
	var zips []File

	osArches := []string{
		"darwin_amd64",
		"linux_amd64",
		"windows_amd64",
	}

	for _, osArch := range osArches {
		full := filepath.Join(path, fmt.Sprintf("%s_%s_%s.zip", project, tag, osArch))
		f, err := os.Open(full)
		if err != nil {
			return nil, err
		}

		zip := File{
			Name: filepath.Base(full),
			Path: filepath.Join(project, tag),
			Type: "application/zip",
			Data: f,
		}

		zips = append(zips, zip)
	}

	return zips, nil
}
