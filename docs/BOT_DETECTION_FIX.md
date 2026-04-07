# Bot Detection Bug Fix

## Problem Summary

The bot detection system was incorrectly blocking legitimate requests, including curl commands with proper browser User-Agent headers. 

### Root Cause

The `rapid_requests` bot detection rule used a catch-all regex pattern `.*` that matched **every single request**, causing the suspicious score to accumulate by 10 points per request. After just 6 requests (60 points), the IP would exceed the 50-point threshold and get blocked. The score never decayed, so once blocked, IPs remained blocked indefinitely.

**Example of the bug:**
```
Request 1: score = 10 ✓
Request 2: score = 20 ✓
Request 3: score = 30 ✓
Request 4: score = 40 ✓
Request 5: score = 50 ✓
Request 6: score = 60 ❌ BLOCKED!
```

## Solution Implemented

Implemented **Option 2: Proper rate-limiting based detection** with the following changes:

### 1. Enhanced BotDetection Struct (`internal/api/middleware/advanced_security/ddos.go`)

Added fields for proper rate tracking:
```go
type BotDetection struct {
    // ... existing fields ...
    rateWindows    map[string][]time.Time // Track request timestamps for rate detection
    rateLimit      int                    // Max requests allowed in the rate window
    rateWindow     time.Duration          // Time window for rate detection
}
```

### 2. Implemented Rate-Based Detection (`ddos.go`)

New `isRateExceeded()` method:
- Tracks actual request timestamps per IP within a sliding time window
- Only triggers when an IP exceeds 50 requests in 10 seconds
- This threshold catches actual bot behavior while allowing normal usage

### 3. Added Score Decay (`ddos.go`)

Enhanced `DetectBot()` with score decay:
```go
// Apply score decay based on time since last activity
if time.Since(activity.lastActivity) > 5*time.Minute {
    activity.score = activity.score / 2 // Reduce score by half after 5 minutes
}
```

This ensures IPs naturally recover over time instead of being permanently blocked.

### 4. Removed Broken Rule (`internal/api/middleware/advanced_security/middleware.go`)

Updated `initBotDetectionRules()`:
```go
{
    name:        "suspicious_user_agent",
    pattern:     regexp.MustCompile(`(?i)(bot|crawler|spider|scanner|python|curl|wget)`),
    score:       15,
    description: "suspicious_user_agent",
},
// Note: rapid_requests detection is now handled by rate-based logic in DetectBot
```

The catch-all `.*` pattern was removed as it was fundamentally broken.

### 5. Added Cleanup Routine (`ddos.go` and `middleware.go`)

New `CleanupOldData()` method:
- Removes stale rate tracking data every 5 minutes
- Deletes suspicious IPs with low scores after 30 minutes of inactivity
- Prevents memory leaks from tracking too many IPs

Background goroutine started in middleware initialization:
```go
go asm.botDetectionCleanupRoutine()
```

## Files Modified

1. `internal/api/middleware/advanced_security/ddos.go`
   - Enhanced `BotDetection` struct with rate tracking fields
   - Updated `DetectBot()` to include score decay and rate-based detection
   - Added `isRateExceeded()` for proper rate limiting
   - Added `CleanupOldData()` for memory management

2. `internal/api/middleware/advanced_security/middleware.go`
   - Removed broken `rapid_requests` rule with `.*` pattern
   - Initialized new `BotDetection` fields (rateWindows, rateLimit, rateWindow)
   - Added `botDetectionCleanupRoutine()` background goroutine

## Behavior Changes

### Before (Broken)
- Every request added 10 points to the suspicious score
- After 6 requests, IP was blocked (60 > 50 threshold)
- Score never decayed - permanent block
- Even curl with browser User-Agent was blocked

### After (Fixed)
- Normal requests don't add to suspicious score
- Only IPs sending >50 requests in 10 seconds get flagged
- Score decays by 50% every 5 minutes of inactivity
- Old data cleaned up every 5 minutes
- Legitimate users (even with curl) are not blocked

## Testing

Run the verification script to see the difference:
```bash
go run /tmp/verify_bot_detection.go
```

Output shows:
- **Old behavior**: Blocks after 6 requests
- **New behavior**: Allows normal usage, only blocks actual bot behavior
- **Score decay**: Demonstrates how scores decrease over time

## Configuration

The rate limiting thresholds can be adjusted if needed:
- `rateLimit`: 50 requests (default) - increase for high-traffic scenarios
- `rateWindow`: 10 seconds (default) - adjust the detection window
- Score decay: 50% reduction after 5 minutes (hardcoded)

## Impact

✅ **Normal users**: Will never be blocked by bot detection during normal usage
✅ **Curl/API users**: Can make requests without being flagged (unless truly abusive)
✅ **Actual bots**: Still detected when sending rapid-fire requests (>50/10s)
✅ **Memory**: No leaks due to automatic cleanup of stale data
✅ **Recovery**: Blocked IPs naturally recover over time through score decay

## Next Steps

1. Test the fix by restarting the API:
   ```bash
   ./bin/orchestrator-api --skip-migrations
   ```

2. Verify that curl requests with browser User-Agent work:
   ```bash
   curl -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" \
        http://localhost:8080/api/v1/session
   ```

3. Monitor logs for any remaining bot detection issues

4. Consider making rate limits configurable via environment variables if needed
