package s3

import (
	"context"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/manifoldco/promulgate/artifact"
)

// Client is a client of s3 for populating our cdn
type Client struct {
	svc    *s3.S3
	bucket string
}

// New creates a new client
func New(bucket string) (*Client, error) {
	if len(bucket) > 4 && bucket[:5] == "s3://" {
		bucket = bucket[5:]
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return &Client{
		svc:    s3.New(sess),
		bucket: bucket,
	}, nil
}

// Put uploads the given file to the s3 bucket
func (c *Client) Put(ctx context.Context, file *artifact.File) error {
	uploader := s3manager.NewUploaderWithClient(c.svc)

	_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket:       aws.String(c.bucket),
		Key:          aws.String(filepath.Join(file.Path, file.Name)),
		Body:         file.Reader(),
		CacheControl: aws.String("public, max-age=604800"),
		ContentType:  aws.String(file.Type),
	})

	return err
}
