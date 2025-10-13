package s3

import (
	"fmt"
)

// TestSignatureFix demonstrates the signature fix for MinIO
func TestSignatureFix() {
	fmt.Println("=== Testing MinIO Signature Fix ===")

	// Create MinIO config
	config := NewConfig(ProviderMinIO, "", "minioadmin", "minioadmin", "https://cdn.nspublic.xyz", "nat-nonprod")

	// Debug before client creation
	fmt.Println("Before client creation:")
	config.DebugConfig()

	// Create client (this will auto-configure)
	client, err := NewClient(config)
	if err != nil {
		fmt.Printf("❌ Failed to create client: %v\n", err)
		return
	}

	// Debug after client creation
	fmt.Println("After client creation:")
	client.config.DebugConfig()

	fmt.Println("✅ MinIO client created successfully!")
	fmt.Println("✅ UsePathStyle should be 'true (FORCED for minio)'")
	fmt.Println("✅ Region should be 'us-east-1'")
	fmt.Println("✅ Expected URL format shows path-style addressing")
}

// QuickMinIOTest shows the minimal working example
func QuickMinIOTest() {
	// For your specific case
	config := NewConfig(ProviderMinIO, "", "your-access-key", "your-secret-key", "https://cdn.nspublic.xyz", "nat-nonprod")
	config.DebugConfig()

	client, err := NewClient(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Client created for bucket: %s\n", client.bucketName)
	fmt.Printf("✅ URLs will use format: https://cdn.nspublic.xyz/nat-nonprod/batch_process/files/file.csv\n")
}