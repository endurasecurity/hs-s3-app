# S3 Self-Signed Certificate Support

## Overview

The AAR Management System S3 client is configured to accept **self-signed certificates** for S3-compatible storage endpoints. This is necessary for:

- **MinIO** with self-signed certificates
- **Local S3-compatible storage** in development/demo environments
- **Private cloud S3 services** with internal certificate authorities

## How It Works

The S3 client (`s3/client.go`) uses a custom HTTP transport with TLS verification disabled:

```go
customHTTPClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // Accepts self-signed certs
        },
    },
}
```

This configuration is applied to all S3 operations:
- ✅ File uploads (PutObject)
- ✅ File downloads (GetObject)
- ✅ Bucket listing (ListObjects)
- ✅ Presigned URLs (GetPresignedURL)

## Security Implications

⚠️ **Warning**: Disabling TLS certificate verification has security implications:

### What This Means

**Disabled**:
- ❌ Certificate authority (CA) validation
- ❌ Hostname verification
- ❌ Certificate expiration checks
- ❌ Protection against MITM attacks with forged certificates

**Still Enabled**:
- ✅ TLS encryption (data is still encrypted in transit)
- ✅ S3 authentication (access key/secret key validation)
- ✅ Bucket permissions

### Risk Assessment

| Environment | Risk Level | Recommendation |
|-------------|------------|----------------|
| **Demo/Testing (Local MinIO)** | Low | ✅ Acceptable |
| **Internal Network (Trusted)** | Low-Medium | ✅ Acceptable with monitoring |
| **Public Internet** | High | ❌ Use proper certificates |
| **Production (AWS S3)** | N/A | Uses valid certs, no issue |

### When This Is Acceptable

✅ **Recommended for:**
- Local development environments
- Security demos and testing
- Isolated/air-gapped networks
- Trusted private cloud deployments
- Temporary proof-of-concept deployments

❌ **NOT recommended for:**
- Production deployments on public networks
- Systems handling real classified data
- Compliance-regulated environments (unless approved)
- Any scenario with untrusted network paths

## Configuration Examples

### Local MinIO with Self-Signed Certs

```bash
# MinIO automatically generates self-signed certs
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

# Application will connect successfully
export S3_ENDPOINT="https://localhost:9000"  # Note: HTTPS with self-signed cert
export S3_ACCESS_KEY="minioadmin"
export S3_SECRET_KEY="minioadmin"
export S3_BUCKET="aar-documents"
```

### AWS S3 (No Impact)

```bash
# AWS S3 uses valid certificates signed by public CAs
# InsecureSkipVerify has no effect - normal validation still occurs
export S3_ENDPOINT=""  # Empty for AWS
export S3_ACCESS_KEY="your-aws-key"
export S3_SECRET_KEY="your-aws-secret"
export S3_BUCKET="your-bucket"
```

## Production Recommendations

If deploying to production with S3-compatible storage, consider these alternatives:

### Option 1: Use Proper Certificates (Best Practice)

```bash
# Install proper CA-signed certificate in MinIO
minio server /data \
  --certs-dir /path/to/certs \
  --console-address ":9001"

# Or use Let's Encrypt for public-facing MinIO
certbot certonly --standalone -d s3.yourdomain.com
```

### Option 2: Add Custom CA to Trust Store

For internal CA-signed certificates:

```go
// Modify s3/client.go to trust specific CA
caCert, _ := os.ReadFile("/path/to/ca.crt")
caCertPool := x509.NewCertPool()
caCertPool.AppendCertsFromPEM(caCert)

customHTTPClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            RootCAs: caCertPool, // Trust specific CA instead of skipping all verification
        },
    },
}
```

### Option 3: Environment Variable Control

Make TLS verification configurable:

```bash
# Add to .env
S3_SKIP_TLS_VERIFY=true  # For dev/demo only
```

```go
// Update s3/client.go
skipVerify := os.Getenv("S3_SKIP_TLS_VERIFY") == "true"
customHTTPClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: skipVerify,
        },
    },
}
```

## Compliance Considerations

### FedRAMP / NIST 800-53

**SC-8: Transmission Confidentiality and Integrity**
- ✅ Encryption: Satisfied (TLS still encrypts)
- ⚠️ Integrity: Weakened (no cert verification)

**SC-13: Cryptographic Protection**
- ✅ Satisfied (TLS 1.2+)

**SC-23: Session Authenticity**
- ⚠️ Weakened without certificate validation

**Recommendation**: Document as accepted risk for demo/dev environments. For production FedRAMP systems, use proper certificates.

### DISA STIGs

**APP3510: Certificate Validation**
- ❌ Finding: Certificate validation disabled
- **Mitigation**: Limit to non-production/demo environments
- **CAT II**: Medium severity

## Troubleshooting

### Certificate Verification Errors (If Re-enabled)

If you re-enable certificate verification and see errors:

```
Error: x509: certificate signed by unknown authority
```

**Solutions**:
1. Install proper CA-signed certificate on S3 endpoint
2. Add CA certificate to system trust store
3. Use custom CA pool (see Option 2 above)

### Connection Refused

```
Error: dial tcp: connect: connection refused
```

This is a network issue, not a certificate issue. Check:
- S3 endpoint is reachable
- Port is correct (9000 for MinIO, 443 for HTTPS)
- Firewall allows connection

## Demo Script Note

When demonstrating to Federal/Defense buyers, acknowledge this configuration:

> "For this demo, we're using MinIO with self-signed certificates, which is common in development and testing environments. The S3 client is configured to accept these certificates. In a production deployment, we would use proper CA-signed certificates or configure the client to trust your organization's internal CA."

## Code Location

The TLS configuration is in:
- **File**: `s3/client.go`
- **Function**: `NewClient()`
- **Lines**: ~36-42

```go
customHTTPClient := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // ← HERE
        },
    },
}
```

---

**Last Updated**: 2025-11-07
**For Demo/Development Use Only**
