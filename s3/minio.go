package s3

// NewConfig creates a universal config with explicit provider
// Usage examples:
//   AWS S3: NewConfig(ProviderAWS, "us-east-1", "key", "secret", "", "bucket")
//   MinIO: NewConfig(ProviderMinIO, "", "key", "secret", "http://localhost:9000", "bucket")
//   DO Spaces: NewConfig(ProviderDigitalOcean, "nyc3", "key", "secret", "https://nyc3.digitaloceanspaces.com", "bucket")
func NewConfig(provider, region, accessKey, secretKey, endpoint, bucketName string) *Config {
	return &Config{
		Provider:        provider,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		BucketName:      bucketName,
		UseSSL:          true, // Default to true, auto-adjusted based on provider
		UsePathStyle:    nil,  // Auto-detect based on provider
	}
}

// NewConfigAuto creates a config that auto-detects provider from endpoint (legacy)
// Usage examples:
//   AWS S3: NewConfigAuto("us-east-1", "key", "secret", "", "bucket")
//   MinIO: NewConfigAuto("", "key", "secret", "http://localhost:9000", "bucket")
//   DO Spaces: NewConfigAuto("nyc3", "key", "secret", "https://nyc3.digitaloceanspaces.com", "bucket")
func NewConfigAuto(region, accessKey, secretKey, endpoint, bucketName string) *Config {
	return &Config{
		Provider:        "", // Will be auto-detected
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		BucketName:      bucketName,
		UseSSL:          true,
		UsePathStyle:    nil,
	}
}

// Legacy helper functions (kept for backward compatibility)

// MinIOConfig creates a config specifically for MinIO
func MinIOConfig(endpoint, accessKey, secretKey, bucketName string) *Config {
	usePathStyle := true
	return &Config{
		Provider:        ProviderMinIO,
		Region:          "us-east-1",
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		BucketName:      bucketName,
		UseSSL:          false,
		UsePathStyle:    &usePathStyle,
	}
}

// MinIOConfigSSL creates a config for MinIO with SSL
func MinIOConfigSSL(endpoint, accessKey, secretKey, bucketName string) *Config {
	usePathStyle := true
	return &Config{
		Provider:        ProviderMinIO,
		Region:          "us-east-1",
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		BucketName:      bucketName,
		UseSSL:          true,
		UsePathStyle:    &usePathStyle,
	}
}

// DigitalOceanSpacesConfig creates a config for DigitalOcean Spaces
func DigitalOceanSpacesConfig(region, accessKey, secretKey, spaceName string) *Config {
	endpoint := "https://" + region + ".digitaloceanspaces.com"
	usePathStyle := false
	return &Config{
		Provider:        ProviderDigitalOcean,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        endpoint,
		BucketName:      spaceName,
		UseSSL:          true,
		UsePathStyle:    &usePathStyle,
	}
}

// AWSS3Config creates a config for AWS S3
func AWSS3Config(region, accessKey, secretKey, bucketName string) *Config {
	usePathStyle := false
	return &Config{
		Provider:        ProviderAWS,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		Endpoint:        "",
		BucketName:      bucketName,
		UseSSL:          true,
		UsePathStyle:    &usePathStyle,
	}
}