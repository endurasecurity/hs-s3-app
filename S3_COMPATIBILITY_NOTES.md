# S3 Compatibility Notes

## Issue: HTTP 500 Errors on Successful Uploads

### Problem Description

Some S3-compatible storage services (particularly certain configurations or versions) return HTTP 500 Internal Server Error responses after successfully uploading files. This manifests as:

```
error uploading file to S3: failed to upload file to S3: operation error S3:
PutObject, exceeded maximum number of attempts, 3, https response error
StatusCode: 500, RequestID: , HostID: , api error InternalServerError:
Internal Server Error
```

**Symptoms:**
- ✅ Files are successfully created in S3 (verified via S3 dashboard/CLI)
- ❌ Application reports upload failure
- ⚠️ AWS SDK retries 3 times (creating the file 3 times, but overwriting same key)
- ⚠️ Application fails to create AAR due to perceived upload failure

### Root Cause

The S3-compatible service has a bug where:
1. File upload succeeds (data is written to storage)
2. Response generation fails (possibly due to metadata processing, logging, etc.)
3. Server returns HTTP 500 instead of HTTP 200 OK
4. AWS SDK treats this as a failure and retries
5. Retries succeed in writing data but fail in response again

This is a **bug in the S3-compatible service**, not in the application or AWS SDK.

### Solution Implemented

The `UploadFile()` method in `s3/client.go` now includes a verification step:

```go
_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{...})

if err != nil {
    // Verify the object actually exists before returning an error
    _, headErr := c.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
        Bucket: aws.String(c.bucketName),
        Key:    aws.String(key),
    })

    if headErr == nil {
        // Object exists despite the error - treat as success
        return nil
    }

    // Object doesn't exist - real failure
    return fmt.Errorf("failed to upload file to S3: %w", err)
}
```

**How It Works:**
1. Attempt PutObject operation
2. If error occurs, verify object exists using HeadObject
3. If object exists → treat as success (ignore the error)
4. If object doesn't exist → real failure (return the error)

### Behavior Changes

**Before Fix:**
```
User uploads AAR with attachment
  → PutObject returns HTTP 500
  → SDK retries 3 times (3 overwrites of same object)
  → Application shows error: "error uploading file to S3"
  → AAR creation fails
  → File exists in S3 but not referenced in application
```

**After Fix:**
```
User uploads AAR with attachment
  → PutObject returns HTTP 500
  → HeadObject verifies file exists
  → Application treats as success
  → AAR creation succeeds
  → File properly linked to AAR in application
```

### Affected Operations

✅ **Fixed:** File uploads via AAR creation form
✅ **Fixed:** All PutObject operations
⚠️ **Not affected:** File downloads (GetObject works fine)
⚠️ **Not affected:** File listing (ListObjects works fine)

### Performance Impact

**Minimal:**
- Successful uploads: No change (no extra request)
- Failed uploads that actually succeeded: +1 HeadObject request (~10-50ms)
- True failures: +1 HeadObject request (negligible compared to failure handling)

**Best case:** 0 extra requests (upload succeeds normally)
**Worst case:** 1 extra request per file (HeadObject verification)

### Alternative Solutions Considered

#### Option 1: Disable Retries
```go
// NOT RECOMMENDED
awsCfg.RetryMaxAttempts = 1
```
**Rejected:** This would disable retries for all operations, including legitimate transient failures.

#### Option 2: Ignore All PutObject Errors
```go
// DANGEROUS - Don't do this
_, err := c.s3Client.PutObject(...)
// Ignore err completely
return nil
```
**Rejected:** This would hide real upload failures, leading to data loss.

#### Option 3: Fix the S3 Service
**Ideal but not always practical:** If you control the S3 service, fix the bug causing HTTP 500 responses.

#### Option 4: Check Error Type
```go
if err != nil {
    if strings.Contains(err.Error(), "InternalServerError") {
        // Check if file exists
    }
}
```
**Rejected:** Too fragile, error messages may vary.

**Our solution (Option 5: Verify Object Exists)** is the most robust.

### Testing the Fix

#### Test 1: Upload New File
```bash
# Create AAR with attachment
# Expected: Success message, file appears in S3
# Status: ✅ PASS
```

#### Test 2: Multiple Files
```bash
# Create AAR with 3 attachments
# Expected: All 3 files upload successfully
# Status: ✅ PASS
```

#### Test 3: True Upload Failure
```bash
# Trigger real failure (invalid credentials, network down, etc.)
# Expected: Proper error message
# Status: ✅ PASS (error still reported correctly)
```

### S3 Service Compatibility

This fix has been tested with:
- ✅ MinIO (various versions with self-signed certs)
- ✅ AWS S3 (no issues, fix is transparent)
- ✅ S3-compatible services returning HTTP 500 on PutObject

### Debugging

If you still see upload failures after this fix:

1. **Check S3 dashboard** - Are files being created?
   - Yes → Issue is fixed, error is elsewhere
   - No → Real upload failure, check credentials/network

2. **Enable detailed logging:**
   ```go
   // In s3/client.go, add logging
   log.Printf("Upload error: %v", err)
   log.Printf("HeadObject result: %v", headErr)
   ```

3. **Verify HeadObject works:**
   ```bash
   aws s3api head-object \
     --bucket your-bucket \
     --key aars/AAR-XXXXX/attachments/file.jpg \
     --endpoint-url https://your-s3-endpoint
   ```

4. **Check S3 service logs** - Look for internal errors

### Monitoring Recommendations

If this workaround is frequently triggered in production:

1. **Log when workaround activates:**
   ```go
   if headErr == nil {
       log.Printf("WARNING: S3 PutObject failed but object exists - " +
                  "S3 service may have a bug (key: %s)", key)
       return nil
   }
   ```

2. **Track frequency:**
   - Frequent triggers → S3 service needs investigation
   - Rare triggers → Acceptable workaround

3. **Report to S3 service vendor** with:
   - S3 service version
   - Request/response headers
   - Frequency of occurrence
   - Example object keys that trigger the issue

### Future Improvements

If you control the S3 service:

1. **Fix the root cause** in the S3 service
2. **Add response logging** to identify what's failing
3. **Upgrade** to a newer version if available
4. **Check configuration** for issues causing 500 errors

### Related Configuration

This fix works in conjunction with:
- ✅ Self-signed certificate support (`InsecureSkipVerify: true`)
- ✅ Path-style addressing (`UsePathStyle: true`)
- ✅ Custom endpoint support

### Code Location

**File:** `s3/client.go`
**Function:** `UploadFile()`
**Lines:** ~95-111

```go
if err != nil {
    // Verification logic here
    _, headErr := c.s3Client.HeadObject(...)
    if headErr == nil {
        return nil // Success despite error
    }
    return fmt.Errorf("failed to upload file to S3: %w", err)
}
```

---

**Last Updated:** 2025-11-07
**Issue Status:** Resolved with workaround
**Long-term Fix:** Requires S3 service vendor to fix HTTP 500 response bug
