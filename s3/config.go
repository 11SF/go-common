package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	BucketName      string
	UseSSL          bool
}

type Client struct {
	s3Client   *s3.Client
	bucketName string
	config     *Config
}

func NewClient(cfg *Config) (*Client, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = &cfg.Endpoint
			o.UsePathStyle = true
		}
	})

	return &Client{
		s3Client:   s3Client,
		bucketName: cfg.BucketName,
		config:     cfg,
	}, nil
}