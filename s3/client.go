package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client wraps the AWS S3 client with application-specific methods
type Client struct {
	s3Client   *s3.Client
	bucketName string
}

// Config holds S3 configuration from environment variables
type Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
}

// NewClient creates a new S3 client with the provided configuration
func NewClient(cfg Config) (*Client, error) {
	// Create custom HTTP client with TLS verification disabled
	// This is necessary for S3-compatible services with self-signed certificates
	customHTTPClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Disable certificate verification for self-signed certs
			},
		},
	}

	// Create custom resolver for custom endpoint (e.g., MinIO, LocalStack)
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		// Use default AWS endpoint
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Load AWS configuration with custom HTTP client
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithHTTPClient(customHTTPClient),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with path-style addressing (required for MinIO and some S3-compatible services)
	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		// Disable strict response validation for S3-compatible services
		// This helps with services that return non-standard S3 responses
		o.DisableS3ExpressSessionAuth = aws.Bool(true)
	})

	return &Client{
		s3Client:   s3Client,
		bucketName: cfg.Bucket,
	}, nil
}

// UploadFile uploads a file to S3 with the specified key
func (c *Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		// Some S3-compatible services return HTTP 500 errors even when the upload succeeds
		// Verify the object actually exists before returning an error
		_, headErr := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(c.bucketName),
			Key:    aws.String(key),
		})

		if headErr == nil {
			// Object exists despite the error - treat as success
			// This is a workaround for S3-compatible services with buggy responses
			return nil
		}

		// Object doesn't exist - real failure
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

// DownloadFile retrieves a file from S3
func (c *Client) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}

// GetPresignedURL generates a presigned URL for temporary access to an S3 object
func (c *Client) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// ListObjects lists all objects under a specific prefix (folder)
func (c *Client) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	result, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in S3: %w", err)
	}

	keys := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		keys = append(keys, *obj.Key)
	}

	return keys, nil
}

// DeleteObject deletes an object from S3
func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

// HeadObject checks if an object exists and returns its metadata
func (c *Client) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	result, err := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	return result, nil
}
