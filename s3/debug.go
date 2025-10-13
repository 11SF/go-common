package s3

import (
	"fmt"
	"log"
)

// DebugConfig prints the current configuration for debugging
func (c *Config) DebugConfig() {
	fmt.Println("=== S3 Configuration Debug ===")
	fmt.Printf("Provider: %s\n", c.Provider)
	fmt.Printf("Region: %s\n", c.Region)
	fmt.Printf("Endpoint: %s\n", c.Endpoint)
	fmt.Printf("BucketName: %s\n", c.BucketName)
	fmt.Printf("UseSSL: %t\n", c.UseSSL)
	if c.UsePathStyle != nil {
		fmt.Printf("UsePathStyle: %t", *c.UsePathStyle)
		if c.Provider == ProviderMinIO || c.Provider == ProviderCustom {
			fmt.Printf(" (FORCED for %s)\n", c.Provider)
		} else {
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("UsePathStyle: <auto-detect>\n")
	}
	fmt.Printf("AccessKeyID: %s\n", maskKey(c.AccessKeyID))
	fmt.Printf("SecretAccessKey: %s\n", maskKey(c.SecretAccessKey))

	// Show expected URL format
	switch c.Provider {
	case ProviderMinIO, ProviderCustom:
		fmt.Printf("Expected URL format: %s/%s/{object-key}\n", c.Endpoint, c.BucketName)
	case ProviderAWS:
		fmt.Printf("Expected URL format: https://%s.s3.%s.amazonaws.com/{object-key}\n", c.BucketName, c.Region)
	case ProviderDigitalOcean:
		fmt.Printf("Expected URL format: %s/{object-key}\n", c.Endpoint)
	}

	fmt.Println("==============================")
}

// LogConfig logs the configuration for debugging (without sensitive data)
func (c *Config) LogConfig() {
	log.Printf("S3 Config - Provider: %s, Region: %s, Endpoint: %s, UseSSL: %t, UsePathStyle: %v",
		c.Provider, c.Region, c.Endpoint, c.UseSSL, c.UsePathStyle)
}

// maskKey masks sensitive key information for logging
func maskKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// ValidateMinIOConfig performs specific validation for MinIO configurations
func ValidateMinIOConfig(cfg *Config) error {
	if cfg.Provider != ProviderMinIO {
		return nil
	}

	if cfg.Endpoint == "" {
		return fmt.Errorf("MinIO requires an endpoint to be specified")
	}

	// MinIO specific checks
	if cfg.UsePathStyle != nil && !*cfg.UsePathStyle {
		log.Printf("WARNING: MinIO should use path-style addressing. Setting UsePathStyle to true")
		usePathStyle := true
		cfg.UsePathStyle = &usePathStyle
	}

	return nil
}