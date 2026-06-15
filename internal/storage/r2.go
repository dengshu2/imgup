package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// R2Options configures an R2 (or any S3-compatible) uploader.
type R2Options struct {
	Endpoint        string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

// R2 is an Uploader backed by Cloudflare R2 via the S3 API.
type R2 struct {
	client *s3.Client
	bucket string
}

// NewR2 builds an R2 uploader. R2 ignores the AWS region but the SDK requires
// one, so we pin "auto" and point the client at the account endpoint with
// virtual-host addressing disabled (R2 uses path-style).
func NewR2(ctx context.Context, o R2Options) (*R2, error) {
	if o.Endpoint == "" || o.Bucket == "" {
		return nil, errors.New("r2: endpoint and bucket are required")
	}
	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion("auto"),
		awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(o.AccessKeyID, o.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(p *s3.Options) {
		p.BaseEndpoint = aws.String(o.Endpoint)
		p.UsePathStyle = true
	})
	return &R2{client: client, bucket: o.Bucket}, nil
}

// Exists reports whether key is already present via a HeadObject call.
func (r *R2) Exists(ctx context.Context, key string) (bool, error) {
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}
	var nf *types.NotFound
	if errors.As(err, &nf) {
		return false, nil
	}
	// HeadObject reports a missing object as a 404 without a typed body in some
	// S3 implementations; treat a generic 404 the same way.
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
		return false, nil
	}
	return false, fmt.Errorf("head %s: %w", key, err)
}

// Put uploads content under key.
func (r *R2) Put(ctx context.Context, key string, content []byte, contentType string) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put %s: %w", key, err)
	}
	return nil
}

// Check confirms the credentials can reach the bucket via HeadBucket.
func (r *R2) Check(ctx context.Context) error {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.bucket),
	})
	if err != nil {
		return fmt.Errorf("head bucket %q: %w", r.bucket, err)
	}
	return nil
}

// List returns objects under prefix, paging through all results unless limit
// caps them.
func (r *R2) List(ctx context.Context, prefix string, limit int) ([]ObjectInfo, error) {
	in := &s3.ListObjectsV2Input{Bucket: aws.String(r.bucket)}
	if prefix != "" {
		in.Prefix = aws.String(prefix)
	}
	p := s3.NewListObjectsV2Paginator(r.client, in)

	var out []ObjectInfo
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list %q: %w", prefix, err)
		}
		for _, o := range page.Contents {
			out = append(out, ObjectInfo{
				Key:          aws.ToString(o.Key),
				Size:         aws.ToInt64(o.Size),
				LastModified: aws.ToTime(o.LastModified),
			})
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
	}
	return out, nil
}

// Delete removes the object at key.
func (r *R2) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete %s: %w", key, err)
	}
	return nil
}
