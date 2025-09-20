package s3

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

func ExampleUsage() {
	// AWS S3 Configuration
	awsConfig := &Config{
		Region:          "us-east-1",
		AccessKeyID:     "your-access-key-id",
		SecretAccessKey: "your-secret-access-key",
		BucketName:      "your-bucket-name",
		UseSSL:          true,
	}

	// MinIO Configuration (uncomment to use)
	/*
	minioConfig := &Config{
		Region:          "us-east-1",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		Endpoint:        "http://localhost:9000",
		BucketName:      "test-bucket",
		UseSSL:          false,
	}
	*/

	// DigitalOcean Spaces Configuration (uncomment to use)
	/*
	doConfig := &Config{
		Region:          "nyc3",
		AccessKeyID:     "your-do-access-key",
		SecretAccessKey: "your-do-secret-key",
		Endpoint:        "https://nyc3.digitaloceanspaces.com",
		BucketName:      "your-space-name",
		UseSSL:          true,
	}
	*/

	// Use any of the configurations above
	client, err := NewClient(awsConfig)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	ctx := context.Background()

	// Upload a file
	data := []byte("Hello, S3!")
	uploadOpts := &UploadOptions{
		ContentType: "text/plain",
		Metadata: map[string]string{
			"author": "go-common",
			"type":   "example",
		},
	}

	err = client.Upload(ctx, "example/hello.txt", data, uploadOpts)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		return
	}
	fmt.Println("✅ File uploaded successfully")

	// Check if file exists
	exists, err := client.Exists(ctx, "example/hello.txt")
	if err != nil {
		log.Printf("Error checking file existence: %v", err)
		return
	}
	fmt.Printf("📁 File exists: %v\n", exists)

	// Get object info
	objInfo, err := client.GetObjectInfo(ctx, "example/hello.txt")
	if err != nil {
		log.Printf("Error getting object info: %v", err)
		return
	}
	fmt.Printf("📊 Object info: Key=%s, Size=%d, LastModified=%v\n",
		objInfo.Key, objInfo.Size, objInfo.LastModified)

	// Download the file
	downloadedData, err := client.Download(ctx, "example/hello.txt")
	if err != nil {
		log.Printf("Download failed: %v", err)
		return
	}
	fmt.Printf("📥 Downloaded content: %s\n", string(downloadedData))

	// List objects with prefix
	objects, err := client.List(ctx, "example/")
	if err != nil {
		log.Printf("List failed: %v", err)
		return
	}
	fmt.Printf("📋 Found %d objects:\n", len(objects))
	for _, obj := range objects {
		fmt.Printf("  - %s (size: %d)\n", obj.Key, obj.Size)
	}

	// Generate presigned URL (valid for 1 hour)
	presignedURL, err := client.GeneratePresignedURL(ctx, "example/hello.txt", time.Hour)
	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
	} else {
		fmt.Printf("🔗 Presigned URL: %s\n", presignedURL)
	}

	// Upload multiple files
	files := map[string][]byte{
		"example/file1.txt": []byte("Content of file 1"),
		"example/file2.txt": []byte("Content of file 2"),
		"example/file3.txt": []byte("Content of file 3"),
	}

	for key, content := range files {
		err = client.Upload(ctx, key, content, nil)
		if err != nil {
			log.Printf("Failed to upload %s: %v", key, err)
		} else {
			fmt.Printf("✅ Uploaded: %s\n", key)
		}
	}

	// Delete multiple files
	keysToDelete := []string{"example/file1.txt", "example/file2.txt", "example/file3.txt"}
	err = client.DeleteMultiple(ctx, keysToDelete)
	if err != nil {
		log.Printf("Failed to delete multiple files: %v", err)
	} else {
		fmt.Printf("🗑️ Deleted %d files\n", len(keysToDelete))
	}

	// Delete the original file
	err = client.Delete(ctx, "example/hello.txt")
	if err != nil {
		log.Printf("Delete failed: %v", err)
		return
	}
	fmt.Println("🗑️ File deleted successfully")
}

func ExampleWithReader() {
	config := &Config{
		Region:          "us-east-1",
		AccessKeyID:     "your-access-key-id",
		SecretAccessKey: "your-secret-access-key",
		BucketName:      "your-bucket-name",
		UseSSL:          true,
	}

	client, err := NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	ctx := context.Background()

	// Upload from reader
	reader := strings.NewReader("Hello from reader!")
	err = client.UploadFromReader(ctx, "example/from-reader.txt", reader, &UploadOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		log.Printf("Upload from reader failed: %v", err)
		return
	}
	fmt.Println("✅ File uploaded from reader successfully")

	// Download to writer (could be a file, buffer, etc.)
	var buffer strings.Builder
	err = client.DownloadToWriter(ctx, "example/from-reader.txt", &buffer)
	if err != nil {
		log.Printf("Download to writer failed: %v", err)
		return
	}
	fmt.Printf("📥 Downloaded to buffer: %s\n", buffer.String())

	// Clean up
	_ = client.Delete(ctx, "example/from-reader.txt")
}