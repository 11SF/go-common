package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
}

type UploadOptions struct {
	ContentType string
	Metadata    map[string]string
}

func (c *Client) Upload(ctx context.Context, key string, data []byte, opts *UploadOptions) error {
	if err := ValidateKey(key); err != nil {
		return err
	}

	if opts == nil {
		opts = &UploadOptions{}
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	if len(opts.Metadata) > 0 {
		input.Metadata = opts.Metadata
	}

	_, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return WrapS3Error(fmt.Errorf("failed to upload object %s: %w", key, err))
	}

	return nil
}

func (c *Client) UploadFromReader(ctx context.Context, key string, reader io.Reader, opts *UploadOptions) error {
	if err := ValidateKey(key); err != nil {
		return err
	}

	if opts == nil {
		opts = &UploadOptions{}
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
		Body:   reader,
	}

	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	if len(opts.Metadata) > 0 {
		input.Metadata = opts.Metadata
	}

	_, err := c.s3Client.PutObject(ctx, input)
	if err != nil {
		return WrapS3Error(fmt.Errorf("failed to upload object %s: %w", key, err))
	}

	return nil
}

func (c *Client) Download(ctx context.Context, key string) ([]byte, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	result, err := c.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, WrapS3Error(fmt.Errorf("failed to download object %s: %w", key, err))
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data %s: %w", key, err)
	}

	return data, nil
}

func (c *Client) DownloadToWriter(ctx context.Context, key string, writer io.Writer) error {
	if err := ValidateKey(key); err != nil {
		return err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	result, err := c.s3Client.GetObject(ctx, input)
	if err != nil {
		return WrapS3Error(fmt.Errorf("failed to download object %s: %w", key, err))
	}
	defer result.Body.Close()

	_, err = io.Copy(writer, result.Body)
	if err != nil {
		return fmt.Errorf("failed to copy object data %s: %w", key, err)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if err := ValidateKey(key); err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return WrapS3Error(fmt.Errorf("failed to delete object %s: %w", key, err))
	}

	return nil
}

func (c *Client) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	var objects []types.ObjectIdentifier
	for _, key := range keys {
		objects = append(objects, types.ObjectIdentifier{
			Key: aws.String(key),
		})
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(c.bucketName),
		Delete: &types.Delete{
			Objects: objects,
		},
	}

	result, err := c.s3Client.DeleteObjects(ctx, input)
	if err != nil {
		return WrapS3Error(fmt.Errorf("failed to delete multiple objects: %w", err))
	}

	if len(result.Errors) > 0 {
		var errorKeys []string
		for _, deleteError := range result.Errors {
			errorKeys = append(errorKeys, aws.ToString(deleteError.Key))
		}
		return fmt.Errorf("failed to delete some objects: %s", strings.Join(errorKeys, ", "))
	}

	return nil
}

func (c *Client) List(ctx context.Context, prefix string) ([]Object, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketName),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	var objects []Object
	paginator := s3.NewListObjectsV2Paginator(c.s3Client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, WrapS3Error(fmt.Errorf("failed to list objects: %w", err))
		}

		for _, obj := range page.Contents {
			objects = append(objects, Object{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				LastModified: aws.ToTime(obj.LastModified),
				ETag:         aws.ToString(obj.ETag),
			})
		}
	}

	return objects, nil
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &notFound) || errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, WrapS3Error(fmt.Errorf("failed to check if object exists %s: %w", key, err))
	}

	return true, nil
}

func (c *Client) GetObjectInfo(ctx context.Context, key string) (*Object, error) {
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	result, err := c.s3Client.HeadObject(ctx, input)
	if err != nil {
		return nil, WrapS3Error(fmt.Errorf("failed to get object info %s: %w", key, err))
	}

	return &Object{
		Key:          key,
		Size:         aws.ToInt64(result.ContentLength),
		LastModified: aws.ToTime(result.LastModified),
		ETag:         aws.ToString(result.ETag),
	}, nil
}

func (c *Client) GeneratePresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	if err := ValidateKey(key); err != nil {
		return "", err
	}

	presignClient := s3.NewPresignClient(c.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})

	if err != nil {
		return "", WrapS3Error(fmt.Errorf("failed to generate presigned URL for %s: %w", key, err))
	}

	return request.URL, nil
}
