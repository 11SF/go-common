package s3

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	ErrObjectNotFound   = errors.New("object not found")
	ErrBucketNotFound   = errors.New("bucket not found")
	ErrAccessDenied     = errors.New("access denied")
	ErrInvalidConfig    = errors.New("invalid configuration")
	ErrEmptyKey         = errors.New("object key cannot be empty")
	ErrEmptyBucketName  = errors.New("bucket name cannot be empty")
)

func WrapS3Error(err error) error {
	if err == nil {
		return nil
	}

	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return fmt.Errorf("%w: %s", ErrObjectNotFound, err.Error())
	}

	var noSuchBucket *types.NoSuchBucket
	if errors.As(err, &noSuchBucket) {
		return fmt.Errorf("%w: %s", ErrBucketNotFound, err.Error())
	}

	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return fmt.Errorf("%w: %s", ErrObjectNotFound, err.Error())
	}

	return err
}

func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return ErrInvalidConfig
	}

	if cfg.BucketName == "" {
		return ErrEmptyBucketName
	}

	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return fmt.Errorf("%w: access key ID and secret access key are required", ErrInvalidConfig)
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	return nil
}

func ValidateKey(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	return nil
}