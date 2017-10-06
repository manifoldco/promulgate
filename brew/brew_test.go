package brew

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/manifoldco/promulgate/artifact"
	"github.com/manifoldco/promulgate/git"
)

func TestNewBottles(t *testing.T) {
	repo := &git.Repository{
		Owner: "manifoldco",
		Name:  "promulgate",
	}

	dir, err := ioutil.TempDir("", "promulgate")

	if err != nil {
		t.Fatalf("Failed to create tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	t.Run("when file is gzip", func(t *testing.T) {
		z, err := os.Create(filepath.Join(dir, "promulgate_0.0.1_darwin_amd64.tar.gz"))
		if err != nil {
			t.Fatalf("Error creating fake zip file %s", err)
		}

		gzw := gzip.NewWriter(z)
		tw := tar.NewWriter(gzw)

		data := []byte("fake binary")
		name := filepath.Join(dir, "promulgate")
		if err := ioutil.WriteFile(name, data, 0644); err != nil {
			t.Fatalf("Failed to create tmp file %s: %s", name, err)
		}

		f, err := os.Open(name)
		if err != nil {
			t.Fatalf("Error reading file %s", err)
		}

		fi, err := f.Stat()
		if err != nil {
			t.Fatalf("Error reading file info %s", err)
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			t.Fatalf("Error creating tar header %s", err)
		}

		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("Error writing tar header %s", err)
		}

		if _, err := io.Copy(tw, f); err != nil {
			t.Fatalf("Error copying file into tar %s", err)
		}

		if err := gzw.Close(); err != nil {
			t.Fatalf("Error closing gzip %s", err)
		}

		if err := tw.Close(); err != nil {
			t.Fatalf("Error closing tar %s", err)
		}

		artifact := &artifact.File{
			Name: "promulgate_0.0.1_darwin_amd64.tar.gz",
			Path: "promulgate/0.0.1",
			Type: "application/gzip",
			Size: fi.Size(),
			Data: z,
		}

		bottles, binname, err := NewBottles(artifact, repo, "0.0.1")
		if err != nil {
			t.Errorf("Error creating bottles %s", err)
		}

		if binname != "promulgate" {
			t.Errorf("Expected bin name to be %s, got %s", "promulgate", binname)
		}

		if len(supportedReleases) != len(bottles) {
			t.Errorf("Expected %d bottle, got %v", len(supportedReleases), len(bottles))
		}
	})

	t.Run("when file is a zip", func(t *testing.T) {
		z, err := os.Create(filepath.Join(dir, "promulgate_0.0.1_darwin_amd64.zip"))
		if err != nil {
			t.Fatalf("Error creating fake zip file %s", err)
		}

		w := zip.NewWriter(z)
		f, err := w.Create("promulgate")
		if err != nil {
			t.Fatalf("Error creating fake file %s", err)
		}
		_, err = f.Write([]byte("fake binary file"))
		if err != nil {
			t.Fatalf("Error saving fake file %s", err)
		}
		err = w.Close()
		if err != nil {
			t.Fatalf("Error saving zip file %s", err)
		}

		info, err := z.Stat()
		if err != nil {
			t.Fatalf("Error reading zip info %s", err)
		}

		artifact := &artifact.File{
			Name: "promulgate_0.0.1_darwin_amd64.zip",
			Path: "promulgate/0.0.1",
			Type: "application/zip",
			Size: info.Size(),
			Data: z,
		}

		bottles, binname, err := NewBottles(artifact, repo, "0.0.1")
		if err != nil {
			t.Errorf("Error creating bottles %s", err)
		}

		if binname != "promulgate" {
			t.Errorf("Expected bin name to be %s, got %s", "promulgate", binname)
		}

		if len(supportedReleases) != len(bottles) {
			t.Errorf("Expected %d bottle, got %v", len(supportedReleases), len(bottles))
		}
	})
}
