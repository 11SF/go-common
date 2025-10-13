package s3

import (
	"context"
	"fmt"
	"log"
)

// ExampleMinIOUsage shows correct MinIO configuration
func ExampleMinIOUsage() {
	// Method 1: Using explicit provider (Recommended)
	config := NewConfig(ProviderMinIO, "", "minioadmin", "minioadmin", "http://localhost:9000", "test-bucket")

	// Method 2: Using helper function
	// config := MinIOConfig("http://localhost:9000", "minioadmin", "minioadmin", "test-bucket")

	// Debug the configuration before creating client
	config.DebugConfig()

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}

	ctx := context.Background()

	// Test basic operations
	fmt.Println("Testing MinIO operations...")

	// Upload test
	data := []byte("Hello MinIO!")
	err = client.Upload(ctx, "test/hello.txt", data, nil)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		return
	}
	fmt.Println("✅ Upload successful")

	// Download test
	downloadedData, err := client.Download(ctx, "test/hello.txt")
	if err != nil {
		log.Printf("Download failed: %v", err)
		return
	}
	fmt.Printf("✅ Download successful: %s\n", string(downloadedData))

	// List test
	objects, err := client.List(ctx, "test/")
	if err != nil {
		log.Printf("List failed: %v", err)
		return
	}
	fmt.Printf("✅ List successful: found %d objects\n", len(objects))

	// Cleanup
	err = client.Delete(ctx, "test/hello.txt")
	if err != nil {
		log.Printf("Delete failed: %v", err)
	} else {
		fmt.Println("✅ Delete successful")
	}
}

// ExampleMinIOSSLUsage shows MinIO with SSL configuration
func ExampleMinIOSSLUsage() {
	// MinIO with SSL (production setup)
	config := NewConfig(ProviderMinIO, "", "your-access-key", "your-secret-key", "https://minio.yourdomain.com", "your-bucket")

	// Or using helper
	// config := MinIOConfigSSL("https://minio.yourdomain.com", "your-access-key", "your-secret-key", "your-bucket")

	config.DebugConfig()

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create MinIO SSL client: %v", err)
	}

	fmt.Println("MinIO SSL client created successfully")

	// Test connection
	ctx := context.Background()
	_, err = client.List(ctx, "")
	if err != nil {
		log.Printf("Failed to list objects: %v", err)
	} else {
		fmt.Println("✅ MinIO SSL connection successful")
	}
}

// TroubleshootMinIO helps debug MinIO connection issues
func TroubleshootMinIO(endpoint, accessKey, secretKey, bucketName string) {
	fmt.Println("=== MinIO Troubleshooting ===")

	// Try auto-detection first
	fmt.Println("1. Testing auto-detection...")
	config1 := NewConfigAuto("", accessKey, secretKey, endpoint, bucketName)
	config1.LogConfig()

	// Try explicit MinIO provider
	fmt.Println("2. Testing explicit MinIO provider...")
	config2 := NewConfig(ProviderMinIO, "", accessKey, secretKey, endpoint, bucketName)
	config2.LogConfig()

	// Try with debugging
	fmt.Println("3. Creating client with debug info...")
	config2.DebugConfig()

	client, err := NewClient(config2)
	if err != nil {
		log.Printf("❌ Client creation failed: %v", err)
		return
	}

	// Test basic connectivity
	ctx := context.Background()
	exists, err := client.Exists(ctx, "non-existent-file")
	if err != nil {
		log.Printf("❌ Connection test failed: %v", err)

		// Common issues and solutions
		fmt.Println("\n=== Common MinIO Issues ===")
		fmt.Println("1. Signature mismatch:")
		fmt.Println("   - Ensure UsePathStyle=true (auto-set for MinIO)")
		fmt.Println("   - Check endpoint URL format (http:// or https://)")
		fmt.Println("   - Verify access key and secret key")
		fmt.Println("2. SSL/TLS issues:")
		fmt.Println("   - Use http:// for local MinIO without SSL")
		fmt.Println("   - Use https:// for production MinIO with SSL")
		fmt.Println("3. Network issues:")
		fmt.Println("   - Check if MinIO server is running")
		fmt.Println("   - Verify endpoint URL is accessible")
		fmt.Println("   - Check firewall/network connectivity")

		return
	}

	fmt.Printf("✅ MinIO connection successful (file exists check: %v)\n", exists)
}