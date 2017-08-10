package s3

import (
	"bytes"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type fileType string

const (
	directoryType fileType = "directory"
	genericType            = "generic"
)

type cdnFile struct {
	Name     string
	Type     fileType
	Size     int64
	Modified time.Time
}

// Listings are sorted lexicographically, with directories coming before
// files.
type sortCDNFile []cdnFile

func (d sortCDNFile) Len() int      { return len(d) }
func (d sortCDNFile) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d sortCDNFile) Less(i, j int) bool {
	if d[i].Type != d[j].Type {
		return d[i].Type == directoryType
	}

	return strings.Compare(d[i].Name, d[j].Name) < 0
}

// CreateIndexes rebuilds index pages for the configured s3 bucket
func (c *Client) CreateIndexes() error {
	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
	}

	listings := make(map[string][]cdnFile)

	err := c.svc.ListObjectsV2Pages(params, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, o := range page.Contents {
			f := cdnFile{
				Name:     filepath.Base(*o.Key),
				Type:     genericType,
				Size:     *o.Size,
				Modified: *o.LastModified,
			}

			// index pages shouldn't list themselves.
			if f.Name == "index.html" {
				continue
			}

			dir := filepath.Dir(*o.Key)
			listings[dir] = append(listings[dir], f)
		}
		return true
	})
	if err != nil {
		return err
	}

	// Add dirs into their parent listings
	for dir := range listings {
		for {
			f := cdnFile{
				Name: filepath.Base(dir),
				Type: directoryType,
			}
			if f.Name == "." {
				break
			}

			dir = filepath.Dir(dir)

			found := false
			for _, o := range listings[dir] {
				if o.Name == f.Name {
					found = true
					break
				}
			}
			if !found {
				listings[dir] = append(listings[dir], f)
			}
		}
	}

	now := time.Now()
	for dir, listing := range listings {
		if dir == "." {
			dir = "/"
		}

		sort.Sort(sortCDNFile(listing))

		buf := &bytes.Buffer{}
		err = tmpl.Execute(buf, struct {
			Dir       string
			Files     []cdnFile
			Timestamp time.Time
		}{Dir: dir, Files: listing, Timestamp: now})

		if err != nil {
			log.Fatal(err)
		}

		dirPath := dir

		params := &s3.PutObjectInput{
			Bucket:       aws.String(c.bucket),
			Key:          aws.String(filepath.Join(dirPath, "index.html")),
			Body:         bytes.NewReader(buf.Bytes()),
			ContentType:  aws.String("text/html"),
			CacheControl: aws.String("public, max-age=300"),
		}
		_, err := c.svc.PutObject(params)
		if err != nil {
			return err
		}
	}

	return nil
}
