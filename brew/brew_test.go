package brew

import (
	"archive/zip"
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
			t.Errorf("Error creating bottes %s", err)
		}

		if binname != "promulgate" {
			t.Errorf("Expected bin name to be %s, got %s", "promulgate", binname)
		}

		if len(supportedReleases) != len(bottles) {
			t.Errorf("Expected %d bottle, got %v", len(supportedReleases), len(bottles))
		}
	})
}
