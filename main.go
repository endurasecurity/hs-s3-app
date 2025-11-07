package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"hs-s3-app/handlers"
	"hs-s3-app/s3"
	"hs-s3-app/storage"
)

func main() {
	// Initialize in-memory storage with pre-populated sample data
	store := storage.NewMemoryStore()
	log.Println("✓ Initialized in-memory storage with sample AARs")

	// S3 is REQUIRED for this demo application
	log.Println("\n═══════════════════════════════════════════════════════════")
	log.Println("  Validating S3 Configuration (REQUIRED)")
	log.Println("═══════════════════════════════════════════════════════════")

	// Check for required S3 environment variables
	s3Config := s3.Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		Region:    getEnvOrDefault("S3_REGION", "us-east-1"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
		Bucket:    os.Getenv("S3_BUCKET"),
	}

	// Validate all required S3 configuration is present
	missingVars := []string{}
	if s3Config.AccessKey == "" {
		missingVars = append(missingVars, "S3_ACCESS_KEY")
	}
	if s3Config.SecretKey == "" {
		missingVars = append(missingVars, "S3_SECRET_KEY")
	}
	if s3Config.Bucket == "" {
		missingVars = append(missingVars, "S3_BUCKET")
	}

	if len(missingVars) > 0 {
		log.Println("\n❌ STARTUP FAILED: Missing required S3 environment variables")
		log.Println("\nMissing variables:")
		for _, v := range missingVars {
			log.Printf("  - %s", v)
		}
		log.Println("\n═══════════════════════════════════════════════════════════")
		log.Println("  S3 Configuration Required")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Println("\nThis application requires S3 storage for the Defense demo.")
		log.Println("\nPlease set the following environment variables:")
		log.Println("\n  export S3_ENDPOINT=\"http://your-s3-endpoint:9000\"")
		log.Println("  export S3_ACCESS_KEY=\"your-access-key\"")
		log.Println("  export S3_SECRET_KEY=\"your-secret-key\"")
		log.Println("  export S3_BUCKET=\"aar-documents\"")
		log.Println("  export S3_REGION=\"us-east-1\"  # Optional, defaults to us-east-1")
		log.Println("\nFor local testing with MinIO:")
		log.Println("\n  # Start MinIO")
		log.Println("  docker run -p 9000:9000 -p 9001:9001 \\")
		log.Println("    -e MINIO_ROOT_USER=minioadmin \\")
		log.Println("    -e MINIO_ROOT_PASSWORD=minioadmin \\")
		log.Println("    minio/minio server /data --console-address \":9001\"")
		log.Println("\n  # Create bucket")
		log.Println("  aws s3 mb s3://aar-documents --endpoint-url http://localhost:9000")
		log.Println("\n  # Set environment variables")
		log.Println("  export S3_ENDPOINT=\"http://localhost:9000\"")
		log.Println("  export S3_ACCESS_KEY=\"minioadmin\"")
		log.Println("  export S3_SECRET_KEY=\"minioadmin\"")
		log.Println("  export S3_BUCKET=\"aar-documents\"")
		log.Println("\nFor AWS S3:")
		log.Println("\n  # Leave S3_ENDPOINT empty for AWS")
		log.Println("  export S3_ENDPOINT=\"\"")
		log.Println("  export S3_ACCESS_KEY=\"your-aws-access-key\"")
		log.Println("  export S3_SECRET_KEY=\"your-aws-secret-key\"")
		log.Println("  export S3_BUCKET=\"your-bucket-name\"")
		log.Println("  export S3_REGION=\"us-east-1\"")
		log.Println("\n═══════════════════════════════════════════════════════════\n")
		os.Exit(1)
	}

	log.Printf("  S3_ENDPOINT:   %s", getDisplayValue(s3Config.Endpoint, "(AWS S3)"))
	log.Printf("  S3_REGION:     %s", s3Config.Region)
	log.Printf("  S3_ACCESS_KEY: %s", maskSecret(s3Config.AccessKey))
	log.Printf("  S3_SECRET_KEY: %s", maskSecret(s3Config.SecretKey))
	log.Printf("  S3_BUCKET:     %s", s3Config.Bucket)

	// Initialize S3 client
	s3Client, err := s3.NewClient(s3Config)
	if err != nil {
		log.Printf("\n❌ STARTUP FAILED: Could not initialize S3 client\n")
		log.Printf("Error: %v\n", err)
		log.Println("\nDiagnostic Information:")
		log.Println("  - Check that S3_ENDPOINT is correct and reachable")
		log.Println("  - For MinIO: Ensure MinIO is running (docker ps)")
		log.Println("  - For AWS S3: Leave S3_ENDPOINT empty or unset")
		log.Println("  - Verify network connectivity to S3 endpoint")
		log.Printf("\nAttempted configuration:")
		log.Printf("  Endpoint: %s\n", getDisplayValue(s3Config.Endpoint, "AWS S3 default"))
		log.Printf("  Region:   %s\n", s3Config.Region)
		log.Printf("  Bucket:   %s\n", s3Config.Bucket)
		log.Println()
		os.Exit(1)
	}

	// Validate S3 connection by checking bucket access
	log.Println("\nValidating S3 connection...")
	if err := validateS3Connection(s3Client, s3Config.Bucket); err != nil {
		log.Printf("\n❌ STARTUP FAILED: Cannot connect to S3 bucket\n")
		log.Printf("Error: %v\n", err)
		log.Println("\nDiagnostic Information:")
		log.Println("  - Verify S3_ACCESS_KEY and S3_SECRET_KEY are correct")
		log.Printf("  - Ensure bucket '%s' exists\n", s3Config.Bucket)
		log.Println("  - Check bucket permissions (ListBucket, GetObject, PutObject)")
		log.Println("  - For MinIO: Verify credentials match MINIO_ROOT_USER/PASSWORD")
		log.Println("  - For AWS S3: Verify IAM credentials and bucket policy")
		log.Println("\nTo create the bucket:")
		if s3Config.Endpoint != "" {
			log.Printf("  aws s3 mb s3://%s --endpoint-url %s\n", s3Config.Bucket, s3Config.Endpoint)
		} else {
			log.Printf("  aws s3 mb s3://%s --region %s\n", s3Config.Bucket, s3Config.Region)
		}
		log.Println()
		os.Exit(1)
	}

	log.Println("✓ S3 client initialized successfully")
	log.Println("✓ S3 bucket connection validated")
	log.Printf("✓ Ready to store documents in: s3://%s\n", s3Config.Bucket)

	// Set up HTTP routes
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Application routes
	mux.HandleFunc("/", handlers.HomeHandler(store))
	mux.HandleFunc("/aar/create", handlers.CreateAARHandler(store, s3Client))
	mux.HandleFunc("/aar/list", handlers.ListAARHandler(store))
	mux.HandleFunc("/aar/view", handlers.ViewAARHandler(store))
	mux.HandleFunc("/aar/download", handlers.DownloadAttachmentHandler(store, s3Client))
	mux.HandleFunc("/aar/generate-report", handlers.GenerateReportHandler(store))

	// Get port from environment or default to 8080
	port := getEnvOrDefault("PORT", "8080")
	addr := fmt.Sprintf(":%s", port)

	// Print startup information
	printBanner(port)

	// Start HTTP server
	log.Printf("Starting HTTP server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// validateS3Connection tests the S3 connection by attempting to list objects in the bucket
func validateS3Connection(client *s3.Client, bucket string) error {
	ctx := context.Background()
	_, err := client.ListObjects(ctx, "")
	return err
}

// maskSecret masks sensitive values for display in logs
func maskSecret(secret string) string {
	if secret == "" {
		return "(not set)"
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// getDisplayValue returns the value or a default display string if empty
func getDisplayValue(value, defaultDisplay string) string {
	if value == "" {
		return defaultDisplay
	}
	return value
}

func printBanner(port string) {
	banner := `
╔═══════════════════════════════════════════════════════════════════════╗
║                                                                       ║
║           ⬢  AFTER ACTION REPORT MANAGEMENT SYSTEM  ⬢                ║
║                                                                       ║
║                    Department of Defense                             ║
║                    UNCLASSIFIED // FOUO                              ║
║                                                                       ║
╚═══════════════════════════════════════════════════════════════════════╝

Server Status: RUNNING
Port: %s
Access URL: http://localhost:%s

Pre-populated Sample AARs:
  • AAR-20251005-0001 - Operation Enduring Shield
  • AAR-20250920-0002 - Exercise Iron Sentinel
  • AAR-20251107-0003 - Operation Phantom Strike (⚠ CONTAINS EXPLOIT PAYLOAD)
  • AAR-20250815-0004 - Exercise Northern Viking

SECURITY WARNING:
This application contains an INTENTIONAL command injection vulnerability
in the PDF report generation feature for runtime security demonstration
purposes. DO NOT deploy this application in production environments.

Vulnerability Location: handlers/aar_report.go
Exploitation: Generate PDF report for "Operation Phantom Strike"

For demo purposes only. Use in controlled environments only.

═══════════════════════════════════════════════════════════════════════
`
	fmt.Printf(banner, port, port)
}
