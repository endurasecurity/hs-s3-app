# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **demonstration application** for runtime security technology demos targeting Federal and Defense sector buyers. It is a Go-based After Action Report (AAR) Management System with an **intentional command injection vulnerability** (RCE) in the PDF generation feature.

**Critical**: This application contains a security vulnerability by design. Never remove or fix the vulnerability in `handlers/aar_report.go` - it exists to demonstrate runtime protection capabilities.

## Build and Run Commands

### Build
```bash
go build -o aar-system
```

### Run
```bash
# Load environment variables from .env
export $(grep -v '^#' .env | xargs)
./aar-system

# Or use the start script (handles .env loading automatically)
./start.sh
```

### Development
```bash
# Run without building
go run main.go

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

## Required Environment Variables

S3 configuration is **mandatory** - the application will not start without it:

```bash
export S3_ENDPOINT="http://localhost:9000"  # Empty for AWS S3
export S3_ACCESS_KEY="minioadmin"
export S3_SECRET_KEY="minioadmin"
export S3_BUCKET="aar-documents"
export S3_REGION="us-east-1"  # Optional, defaults to us-east-1
```

The application validates S3 connectivity on startup and provides detailed diagnostic output if configuration is missing or invalid.

## Architecture

### Request Flow
```
HTTP Request → main.go (routing) → handlers/*.go → models/storage → Response
                                         ↓
                                    s3/client.go (for file operations)
```

### Key Components

**main.go**
- Entry point and HTTP server setup
- **S3 validation on startup** - fails fast with diagnostics if S3 is misconfigured
- Routes all requests to appropriate handlers
- No authentication/authorization (demo purposes)

**handlers/**
- Each handler returns `http.HandlerFunc` that wraps dependencies (store, s3Client)
- `aar_report.go` contains the **intentional RCE vulnerability** at line ~57
  - Vulnerability: `aar.OperationName` is directly interpolated into shell command
  - Exploitation vector: `Operation Name'; malicious_command; echo 'suffix`

**models/aar.go**
- Single data model with Defense/Military-specific fields
- Includes AAR metadata and Attachment records

**storage/memory.go**
- In-memory storage (no database)
- Pre-populated with 4 sample AARs on initialization
- AAR-20251107-0003 ("Operation Phantom Strike") contains exploit payload

**s3/client.go**
- Wraps AWS SDK for Go v2
- **Accepts self-signed certificates** (`InsecureSkipVerify: true`)
- **Workaround for HTTP 500 errors**: If `PutObject` fails, verifies object exists with `HeadObject` before reporting failure
- This handles buggy S3-compatible services that return 500 after successful uploads

**templates/**
- Server-side rendered HTML using Go's `html/template`
- `layout.html` provides base structure with DoD classification banners
- All pages use Defense/Military styling (navy blue #002F6C)

## Important Implementation Details

### S3 Configuration Workarounds

1. **Self-Signed Certificates**: The S3 client disables TLS verification to support MinIO and other services with self-signed certs. See `S3_SELF_SIGNED_CERTS.md` for security implications.

2. **HTTP 500 Upload Workaround**: Some S3-compatible services return HTTP 500 after successfully uploading files. The `UploadFile()` method verifies object existence before treating as failure. See `S3_COMPATIBILITY_NOTES.md` for details.

### The Intentional Vulnerability

**Location**: `handlers/aar_report.go:57`

**Code**:
```go
cmdStr := fmt.Sprintf("wkhtmltopdf --title '%s' %s %s",
    aar.OperationName,  // VULNERABLE: User-controlled, unsanitized
    tmpHTMLFile,
    tmpPDFFile)
cmd := exec.Command("sh", "-c", cmdStr)
```

**Purpose**: Demonstrates command injection for runtime security demos. Pre-populated AAR "Operation Phantom Strike" contains a ready-to-exploit payload.

**Do NOT**:
- Fix or sanitize this vulnerability
- Add input validation to `OperationName`
- Remove the `sh -c` command wrapper
- Change the command construction method

### Template Function Registration

When working with templates that use custom functions (e.g., `detail.html` uses `divf`), the function must be registered in the handler:

```go
funcMap := template.FuncMap{
    "divf": func(a int64, b float64) float64 {
        return float64(a) / b
    },
}
tmpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(...)
```

## File Upload Flow

1. User submits AAR creation form with multipart file uploads
2. `handlers/aar_create.go` processes form
3. Files stream directly to S3 via `s3Client.UploadFile()`
4. S3 keys follow pattern: `aars/{aar-id}/attachments/{filename}`
5. Attachment metadata stored in AAR's `Attachments` slice
6. Pre-populated AARs have mock attachment metadata but files don't exist in S3

## Development Notes

### Adding New AAR Fields
1. Update `models/aar.go` struct
2. Add form field in `templates/create.html`
3. Update form parsing in `handlers/aar_create.go`
4. Update display in `templates/detail.html` and `templates/list.html`

### Modifying Pre-Populated Data
Edit `storage/memory.go` `loadSampleData()` function. AAR-20251107-0003 must retain the exploit payload in `OperationName`.

### Changing S3 Behavior
All S3 operations go through `s3/client.go`. The client is configured for:
- Path-style addressing (MinIO compatibility)
- Self-signed certificate acceptance
- Upload verification workaround

## Documentation Reference

- **README.md**: Comprehensive user documentation, setup, vulnerability details
- **QUICKSTART.md**: 5-minute setup guide
- **DEMO_SCRIPT.md**: Full presentation script for Federal/Defense buyers (15-20 min)
- **EXPLOIT_PAYLOADS.md**: 50+ exploitation examples and attack listener setup
- **S3_SELF_SIGNED_CERTS.md**: TLS verification details and security implications
- **S3_COMPATIBILITY_NOTES.md**: HTTP 500 workaround explanation

## External Dependencies

- **wkhtmltopdf**: Required for PDF generation (and vulnerability exploitation)
  ```bash
  # Ubuntu/Debian
  sudo apt-get install wkhtmltopdf

  # macOS
  brew install wkhtmltopdf
  ```

- **S3-compatible storage**: MinIO (recommended for local dev), AWS S3, or compatible service
  ```bash
  # Start MinIO locally
  docker run -p 9000:9000 -p 9001:9001 \
    -e MINIO_ROOT_USER=minioadmin \
    -e MINIO_ROOT_PASSWORD=minioadmin \
    minio/minio server /data --console-address ":9001"

  # Create bucket
  aws s3 mb s3://aar-documents --endpoint-url http://localhost:9000
  ```

## When to Reference Other Docs

- **Adding exploit payloads**: See `EXPLOIT_PAYLOADS.md`
- **S3 connectivity issues**: Check `S3_COMPATIBILITY_NOTES.md` and startup diagnostics
- **Certificate errors**: See `S3_SELF_SIGNED_CERTS.md`
- **Presenting to buyers**: Use `DEMO_SCRIPT.md` as script template
- **Understanding vulnerability**: See README.md "The Vulnerability" section

## Project Constraints

1. **No authentication/authorization**: Demo application, intentionally simplified
2. **No database**: In-memory storage only, data lost on restart
3. **No tests**: Demonstration code, not production-quality
4. **Vulnerability must remain**: This is the core feature for security demos
5. **S3 is mandatory**: Application design centers on S3 document storage demo
