package artifact

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Release represents a release object
type Release struct {
	Tag  string // tag
	Body string // markdown formatted release notes
}

// Multireader exposes many reader interfaces
type Multireader interface {
	io.ReadSeeker
	io.ReaderAt
}

// File represents a file on disk
type File struct {
	Name string
	Path string
	Type string

	Size int64
	Data Multireader
}

// Reader returns an io.Reader for the file.
func (f *File) Reader() io.Reader {
	f.Data.Seek(0, io.SeekStart)
	return f.Data
}

// ReaderAt returns an io.ReaderAt for the file.
func (f *File) ReaderAt() io.ReaderAt {
	return f.Data
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

		s, err := f.Stat()
		if err != nil {
			return nil, err
		}

		zip := File{
			Name: filepath.Base(full),
			Path: filepath.Join(project, tag),
			Type: "application/zip",
			Size: s.Size(),
			Data: f,
		}

		zips = append(zips, zip)
	}

	return zips, nil
}

// Sha256 returns the hex encoded
func (f *File) Sha256() (string, error) {
	b, err := ioutil.ReadAll(f.Reader())
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return fmt.Sprintf("%x", sum), nil
}
