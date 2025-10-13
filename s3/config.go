package s3

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Provider constants
const (
	ProviderAWS          = "aws"
	ProviderMinIO        = "minio"
	ProviderDigitalOcean = "digitalocean"
	ProviderCustom       = "custom"
)

type Config struct {
	Provider        string // Provider type: aws, minio, digitalocean, custom
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string // Empty for AWS S3, set for MinIO/DigitalOcean Spaces
	BucketName      string
	UseSSL          bool
	UsePathStyle    *bool // Optional: auto-detect if nil based on provider
}

type Client struct {
	s3Client   *s3.Client
	bucketName string
	config     *Config
}

// validateProvider checks if the provider is valid
func validateProvider(provider string) error {
	switch provider {
	case ProviderAWS, ProviderMinIO, ProviderDigitalOcean, ProviderCustom:
		return nil
	default:
		return fmt.Errorf("invalid provider: %s. Valid providers are: %s, %s, %s, %s",
			provider, ProviderAWS, ProviderMinIO, ProviderDigitalOcean, ProviderCustom)
	}
}

// detectProviderType determines the S3 provider based on endpoint (fallback only)
func detectProviderType(endpoint string) string {
	if endpoint == "" {
		return ProviderAWS
	}

	endpoint = strings.ToLower(endpoint)

	if strings.Contains(endpoint, "digitaloceanspaces.com") {
		return ProviderDigitalOcean
	}

	// Common MinIO patterns
	if strings.Contains(endpoint, "localhost") ||
	   strings.Contains(endpoint, "127.0.0.1") ||
	   strings.Contains(endpoint, ":9000") ||
	   strings.Contains(endpoint, "minio") {
		return ProviderMinIO
	}

	// Default to custom for other endpoints
	return ProviderCustom
}

// autoConfigureForProvider sets optimal defaults for each provider
func autoConfigureForProvider(cfg *Config) {
	// Use explicit provider if set, otherwise auto-detect
	providerType := cfg.Provider
	if providerType == "" {
		providerType = detectProviderType(cfg.Endpoint)
		cfg.Provider = providerType
	}

	// Set default region if empty
	if cfg.Region == "" {
		switch providerType {
		case ProviderAWS:
			cfg.Region = "us-east-1"
		case ProviderDigitalOcean:
			cfg.Region = "nyc3" // default DO region
		default: // MinIO and custom
			cfg.Region = "us-east-1" // MinIO doesn't care but SDK requires it
		}
	}

	// Force us-east-1 for MinIO if region is empty or different (signature compatibility)
	if providerType == ProviderMinIO && cfg.Region != "us-east-1" {
		cfg.Region = "us-east-1"
	}

	// Auto-configure SSL based on endpoint scheme for MinIO
	if providerType == ProviderMinIO || providerType == ProviderCustom {
		if cfg.Endpoint != "" {
			if strings.HasPrefix(strings.ToLower(cfg.Endpoint), "http://") {
				cfg.UseSSL = false
			} else if strings.HasPrefix(strings.ToLower(cfg.Endpoint), "https://") {
				cfg.UseSSL = true
			}
		}
	}

	// Auto-configure UsePathStyle if not explicitly set
	if cfg.UsePathStyle == nil {
		switch providerType {
		case ProviderAWS:
			usePathStyle := false // AWS S3 prefers virtual hosted-style
			cfg.UsePathStyle = &usePathStyle
		case ProviderDigitalOcean:
			usePathStyle := false // DO Spaces supports virtual hosted-style
			cfg.UsePathStyle = &usePathStyle
		default: // MinIO and custom S3-compatible
			usePathStyle := true // Must use path-style for proper signatures
			cfg.UsePathStyle = &usePathStyle
		}
	}

	// Force UsePathStyle=true for MinIO regardless of user setting (critical for signature)
	if providerType == ProviderMinIO || providerType == ProviderCustom {
		usePathStyle := true
		cfg.UsePathStyle = &usePathStyle
	}
}

func NewClient(cfg *Config) (*Client, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	// Validate provider if explicitly set
	if cfg.Provider != "" {
		if err := validateProvider(cfg.Provider); err != nil {
			return nil, err
		}
	}

	// Auto-configure based on provider
	autoConfigureForProvider(cfg)

	// MinIO specific validation
	if err := ValidateMinIOConfig(cfg); err != nil {
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
		}

		// Force path style for MinIO and custom providers (CRITICAL for signature)
		if cfg.Provider == ProviderMinIO || cfg.Provider == ProviderCustom {
			o.UsePathStyle = true
		} else if cfg.UsePathStyle != nil && *cfg.UsePathStyle {
			o.UsePathStyle = true
		}
	})

	return &Client{
		s3Client:   s3Client,
		bucketName: cfg.BucketName,
		config:     cfg,
	}, nil
}