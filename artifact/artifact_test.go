package artifact

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestFindCompressedFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "promulgate")

	if err != nil {
		t.Fatalf("Failed to create tmp dir: %s", err)
	}

	defer os.RemoveAll(dir)

	type example struct {
		mimetype string
		include  bool
	}

	tcs := map[string]example{
		"promulgate_0.0.1_linux_amd64.tar.gz":     {mimetype: Gzip, include: true},
		"promulgate_0.0.1_linux_amd64.zip":        {mimetype: Zip, include: true},
		"promulgate_0.0.1_darwin_amd64.tar.gz":    {mimetype: Gzip, include: true},
		"promulgate_0.0.1_darwin_amd64.zip":       {mimetype: Zip, include: true},
		"promulgate_0.0.1_windows_amd64.zip":      {mimetype: Zip, include: true},
		"promulgate_0.0.2_linux_amd64.tar.gz":     {include: false},
		"not_promulgate_0.0.1_linux_amd64.tar.gz": {include: false},
	}

	for name := range tcs {
		data := []byte("tmp file")

		name := filepath.Join(dir, name)
		if err := ioutil.WriteFile(name, data, 0644); err != nil {
			t.Fatalf("Failed to create tmp file %s: %s", name, err)
		}
	}

	files, err := FindCompressedFiles(dir, "promulgate", "0.0.1")

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

	for _, f := range files {
		eg, ok := tcs[f.Name]

		if !ok {
			t.Errorf("Example file not found %s", f.Name)
		}

		if !eg.include {
			t.Errorf("Expected file not to be included %s", f.Name)
		}

		if eg.mimetype != f.Type {
			t.Errorf("Wrong mimetype, expected %s, got %s", eg.mimetype, f.Name)
		}
	}
}
