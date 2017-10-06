package artifact

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const Zip = "application/zip"
const Gzip = "application/gzip"

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

// FindCompressedFiles finds the pre-built zips for the given tag in the given directory
func FindCompressedFiles(path, project, tag string) ([]File, error) {
	var files []File

	osArches := []string{
		"darwin_amd64",
		"linux_amd64",
		"windows_amd64",
	}

	formats := []struct {
		extension string
		mimetype  string
	}{
		{"zip", Zip},
		{"tar.gz", Gzip},
	}

	for _, osArch := range osArches {
		for _, format := range formats {
			name := fmt.Sprintf("%s_%s_%s.%s", project, tag, osArch, format.extension)
			full := filepath.Join(path, name)

			f, err := os.Open(full)
			if os.IsNotExist(err) {
				continue
			}

			if err != nil {
				return nil, err
			}

			s, err := f.Stat()
			if err != nil {
				return nil, err
			}

			file := File{
				Name: filepath.Base(full),
				Path: filepath.Join(project, tag),
				Type: format.mimetype,
				Size: s.Size(),
				Data: f,
			}

			files = append(files, file)
		}
	}

	return files, nil
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
