package brew

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promulgate/artifact"
	"github.com/manifoldco/promulgate/git"
)

var supportedReleases = []string{"high_sierra", "sierra", "el_capitan", "yosemite"}

// NewBottles converts the provided zip file into a tarball suitable for use as
// a brew bottle.
func NewBottles(darwin *artifact.File, repo *git.Repository, tag string) ([]artifact.File, string, error) {
	var file io.Reader
	var fi os.FileInfo
	var err error

	switch t := darwin.Type; t {
	case artifact.Gzip:
		file, fi, err = openGzip(darwin)
	case artifact.Zip:
		file, fi, err = openZip(darwin)
	default:
		err = fmt.Errorf("File format %s unknown", t)
	}

	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer gz.Close()
	tgz := tar.NewWriter(gz)
	defer tgz.Close()

	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return nil, "", err
	}

	parts := strings.SplitN(darwin.Name, "_", -1)

	binname := hdr.Name
	hdr.Name = filepath.Join(parts[0], parts[1], "bin", hdr.Name)
	err = tgz.WriteHeader(hdr)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(tgz, file)
	if err != nil {
		return nil, "", err
	}

	var jbuf bytes.Buffer
	err = receipt.Execute(&jbuf, receiptArgs{
		Owner: repo.Owner,
		Name:  repo.Name,
		Tag:   parts[1],
		Time:  time.Now().Unix(),
	})
	if err != nil {
		return nil, "", err
	}

	hdr = &tar.Header{
		Name: filepath.Join(parts[0], parts[1], "INSTALL_RECEIPT.json"),
		Mode: 0600,
		Size: int64(jbuf.Len()),
	}
	err = tgz.WriteHeader(hdr)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(tgz, &jbuf)
	if err != nil {
		return nil, "", err
	}

	tgz.Close()
	gz.Close()

	var bottles []artifact.File

	for _, rel := range supportedReleases {
		bottles = append(bottles, artifact.File{
			Name: fmt.Sprintf("%s-%s.%s.bottle.tar.gz", parts[0], parts[1], rel),
			Path: filepath.Join(parts[0], "brew", "bottles"),
			Type: "application/x-tar",
			Size: int64(buf.Len()),
			Data: bytes.NewReader(buf.Bytes()),
		})
	}
	return bottles, binname, nil
}

func openGzip(darwin *artifact.File) (io.Reader, os.FileInfo, error) {
	gz, err := gzip.NewReader(darwin.Reader())
	if err != nil {
		return nil, nil, err
	}

	r := tar.NewReader(gz)

	for {
		header, err := r.Next()
		switch {
		case err == io.EOF:
			return nil, nil, err
		case err != nil:
			return nil, nil, err
		case header == nil:
			continue
		}

		var buf bytes.Buffer

		_, err = io.Copy(&buf, r)
		if err != nil && err != io.EOF {
			return nil, nil, err
		}

		return &buf, header.FileInfo(), nil
	}
}

func openZip(darwin *artifact.File) (io.Reader, os.FileInfo, error) {
	zr, err := zip.NewReader(darwin.ReaderAt(), darwin.Size)
	if err != nil {
		return nil, nil, err
	}

	if len(zr.File) != 1 {
		return nil, nil, errors.New("zip file must contain only 1 file")
	}

	fi := zr.File[0].FileInfo()

	zf, err := zr.File[0].Open()
	if err != nil {
		return nil, nil, err
	}

	return zf, fi, err
}

// NewFormula returns a file whose contents are a valid brew formula
func NewFormula(repo *git.Repository, tag, binname, homepage, description, cdn string, bottles []artifact.File) (*artifact.File, error) {

	sum, err := bottles[0].Sha256()
	if err != nil {
		return nil, err
	}

	var bottleSums []formulaBottleArgs

	for _, bottle := range bottles {
		parts := strings.SplitN(bottle.Name, ".", -1)
		bottleSums = append(bottleSums, formulaBottleArgs{
			Checksum: sum,
			Name:     parts[len(parts)-4],
		})
	}

	var buf bytes.Buffer
	err = formula.Execute(&buf, formulaArgs{
		FormulaName: strings.Replace(strings.Title(repo.Name), "-", "", -1),
		Owner:       repo.Owner,
		Name:        repo.Name,
		Tag:         tag,
		Description: description,
		Homepage:    homepage,

		BinName: binname,

		Checksum: "",

		BottleURL: cdn + repo.Name + "/brew/bottles",
		Bottles:   bottleSums,
	})
	if err != nil {
		return nil, err
	}

	return &artifact.File{
		Name: repo.Name + ".rb",
		Path: "Formula",
		Type: "text/plain",
		Size: int64(buf.Len()),
		Data: bytes.NewReader(buf.Bytes()),
	}, nil
}
