# SubSpace Troubleshooting Guide

## Common Issues and Solutions

### Installation Issues

#### Issue: "go: command not found"
**Solution**: Install Go from https://go.dev/dl/

```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

#### Issue: "go.mod: no such file or directory"
**Solution**: The go.mod file should be in the project root. Verify you're in the correct directory:

```bash
cd /path/to/SubSpace
ls go.mod  # Should exist
```

#### Issue: Dependency download fails
**Solution**: 

```bash
# Clear Go module cache
go clean -modcache

# Re-download dependencies
go mod download

# If behind a proxy
export GOPROXY=https://proxy.golang.org,direct
go mod download
```

---

### Build Issues

#### Issue: "package github.com/saurabhkuntal/subspace/pkg/... is not in GOROOT"
**Solution**: Module path doesn't match. The module is defined in go.mod:

```bash
# Verify module name in go.mod
head -1 go.mod
# Should show: module github.com/saurabhkuntal/subspace

# Rebuild
go build -o subspace main.go
```

#### Issue: Build succeeds but binary doesn't run
**Solution**: Check permissions and architecture:

```bash
# Make executable
chmod +x subspace

# Check architecture
file subspace

# Run directly
./subspace -help
```

---

### Configuration Issues

#### Issue: "Failed to load configuration: LinkedIn email is required"
**Solution**: Create and configure .env file:

```bash
# Copy template
cp .env.example .env

# Edit with your credentials
nano .env

# Verify
cat .env | grep LINKEDIN_EMAIL
```

#### Issue: "invalid configuration: max action delay must be greater than min"
**Solution**: Check delay settings in .env:

```env
MIN_ACTION_DELAY=2000
MAX_ACTION_DELAY=5000  # Must be >= MIN_ACTION_DELAY
```

---

### Authentication Issues

#### Issue: "Failed to login: failed to find email input"
**Solution**: LinkedIn's UI may have changed. Try with visible browser:

```bash
HEADLESS=false ./subspace -action=search
```

If the login page looks different, selectors may need updating in `pkg/auth/auth.go`.

#### Issue: "2FA or security verification required"
**Solution**: LinkedIn detected automation or requires verification:

1. **Manual verification**: Complete 2FA manually in visible browser
2. **Use test account**: Create a dedicated test account
3. **Wait period**: Wait 24-48 hours before retrying

#### Issue: "Session expired, performing fresh login" every time
**Solution**: Session cookies aren't being saved:

```bash
# Check session directory
ls -la data/sessions/

# Ensure write permissions
chmod 755 data/sessions

# Clear and retry
rm -rf data/sessions/*
./subspace -clear-session -action=search
```

#### Issue: "Login failed - invalid credentials"
**Solution**:

1. Verify credentials in .env are correct
2. Try logging in manually on LinkedIn.com
3. Check for special characters that need escaping
4. Ensure no trailing spaces in .env values

---

### Browser Issues

#### Issue: "Failed to connect to browser"
**Solution**:

```bash
# Install Chrome/Chromium
# macOS
brew install --cask google-chrome

# Linux
sudo apt install chromium-browser

# Verify Chrome is installed
which google-chrome
```

#### Issue: Browser opens but is blank/frozen
**Solution**: 

```bash
# Try without headless mode
HEADLESS=false ./subspace -action=search

# Add slow motion for visibility
SLOW_MOTION=500 ./subspace -action=search

# Check for conflicting browser processes
ps aux | grep chrome
killall chrome  # If needed
```

#### Issue: "Failed to apply stealth settings"
**Solution**: Rod stealth plugin issue:

```bash
# Update dependencies
go get -u github.com/go-rod/stealth
go mod tidy

# Rebuild
go build -o subspace main.go
```

---

### Search Issues

#### Issue: "Search results not found"
**Solution**: LinkedIn's search page structure changed:

```bash
# Run with debug logging
LOG_LEVEL=debug ./subspace -action=search -keywords="test"

# Try with visible browser to see what's happening
HEADLESS=false LOG_LEVEL=debug ./subspace -action=search
```

Update selectors in `pkg/search/search.go` if needed.

#### Issue: "Found 0 results" but there should be results
**Solution**:

1. **Check search criteria**: Too specific keywords
2. **Location issues**: Location filter not matching
3. **Network filter**: Only searches 2nd/3rd connections
4. **Try manual search**: Open LinkedIn and try same search

```bash
# Try broader search
./subspace -action=search -keywords="Engineer" -location="United States"
```

---

### Connection Request Issues

#### Issue: "Failed to find connect button"
**Solution**: Profile may already be connected or pending:

```bash
# Check database for existing requests
sqlite3 data/subspace.db "SELECT * FROM connection_requests WHERE profile_url LIKE '%username%';"

# Or the person has connection limits enabled
```

#### Issue: "Reached daily limit" but it's a new day
**Solution**: Database tracks by 24-hour period, not calendar day:

```bash
# Check current count
sqlite3 data/subspace.db "SELECT COUNT(*) FROM connection_requests WHERE sent_at >= datetime('now', '-1 day');"

# Manually reset if needed (CAREFUL!)
sqlite3 data/subspace.db "DELETE FROM connection_requests WHERE sent_at < datetime('now', '-7 days');"
```

#### Issue: Connection note is truncated
**Solution**: LinkedIn limits notes to 300 characters. The code automatically truncates, but verify:

```bash
# Message should be <= 300 chars
./subspace -action=connect -message="$(head -c 300 <<< 'Your long message...')"
```

---

### Messaging Issues

#### Issue: "Failed to find message input"
**Solution**: LinkedIn messaging UI varies by profile type:

```bash
# Try with visible browser
HEADLESS=false ./subspace -action=message

# Check if profiles are actually connected
./subspace -action=detect-connections
```

#### Issue: "No new connections found" but you accepted some
**Solution**: Detection logic checks for absence of "Connect" button:

```bash
# Manually verify in database
sqlite3 data/subspace.db "SELECT * FROM connection_requests WHERE status='pending';"

# Update status manually if needed
sqlite3 data/subspace.db "UPDATE connection_requests SET status='accepted' WHERE profile_url='...';"
```

---

### Database Issues

#### Issue: "Failed to open database: unable to open database file"
**Solution**: Permission or path issue:

```bash
# Create data directory
mkdir -p data

# Check permissions
ls -ld data/

# Ensure write access
chmod 755 data/

# Try relative path
DATABASE_PATH=./data/subspace.db ./subspace -action=search
```

#### Issue: "Database is locked"
**Solution**: Another process is using the database:

```bash
# Check for running instances
ps aux | grep subspace

# Kill if needed
killall subspace

# Remove lock file
rm -f data/subspace.db-shm data/subspace.db-wal
```

#### Issue: Want to reset database
**Solution**:

```bash
# Backup first
cp data/subspace.db data/subspace.db.backup

# Remove database
rm data/subspace.db

# Will be recreated on next run
./subspace -action=search
```

---

### Rate Limiting Issues

#### Issue: Hitting rate limits too quickly
**Solution**: Adjust limits in .env:

```env
# More conservative limits
MAX_CONNECTIONS_PER_DAY=10
MAX_CONNECTIONS_PER_HOUR=2
MAX_MESSAGES_PER_DAY=8

# Longer delays
MIN_ACTION_DELAY=5000
MAX_ACTION_DELAY=10000
```

#### Issue: Want to bypass rate limits for testing
**Solution**: **NOT RECOMMENDED** but for testing:

```env
# Set very high limits
MAX_CONNECTIONS_PER_DAY=1000
MAX_CONNECTIONS_PER_HOUR=100

# Or manually clear database
sqlite3 data/subspace.db "DELETE FROM connection_requests;"
```

---

### Performance Issues

#### Issue: Application is very slow
**Solution**:

```bash
# Reduce stealth delays for faster testing
MIN_ACTION_DELAY=500
MAX_ACTION_DELAY=1000

# Disable some stealth features
ENABLE_RANDOM_SCROLL=false
ENABLE_TYPING_SIMULATION=false

# Use headless mode
HEADLESS=true
```

#### Issue: High memory usage
**Solution**:

```bash
# Limit search results
./subspace -action=search -max-results=5

# Close browser between operations
# (Already handled in code)

# Monitor memory
watch -n 1 'ps aux | grep subspace'
```

---

### Detection & Bans

#### Issue: "Your account has been restricted"
**Solution**: LinkedIn detected automation:

1. **Stop immediately**: Don't continue using tool
2. **Appeal**: Contact LinkedIn support
3. **Wait period**: 1-2 weeks before appeal
4. **Use test account**: For future development
5. **Review limits**: Were you too aggressive?

**Prevention**:
- Use conservative rate limits
- Enable all stealth features
- Operate during business hours only
- Use realistic delays
- Don't automate on main account

#### Issue: Captcha appearing frequently
**Solution**: Heightened detection:

1. **Reduce frequency**: Lower daily/hourly limits
2. **Longer delays**: Increase action delays
3. **Operating hours**: Respect business hours
4. **Break patterns**: Enable break simulation
5. **Account age**: Older accounts less suspicious

---

### Debugging Tips

#### Enable Maximum Logging

```bash
LOG_LEVEL=debug HEADLESS=false SLOW_MOTION=1000 ./subspace -action=search
```

#### Inspect Database

```bash
# Open SQLite shell
sqlite3 data/subspace.db

# Useful queries
SELECT * FROM connection_requests ORDER BY sent_at DESC LIMIT 10;
SELECT COUNT(*) FROM connection_requests WHERE DATE(sent_at) = DATE('now');
SELECT * FROM messages ORDER BY sent_at DESC;
SELECT * FROM search_history;

# Exit
.exit
```

#### Check Browser Console

When running in visible mode, open browser DevTools (F12) to see:
- Network requests
- Console errors
- Element selection issues

#### Test Individual Functions

Create a test script:

```go
package main

import (
    "github.com/saurabhkuntal/subspace/pkg/stealth"
    // ... test individual packages
)
```

---

### Getting Help

1. **Check Documentation**:
   - README.md - Overview and usage
   - DEVELOPMENT.md - Technical details
   - This file - Troubleshooting

2. **Enable Debug Logging**:
   ```bash
   LOG_LEVEL=debug ./subspace -action=search 2>&1 | tee debug.log
   ```

3. **Check GitHub Issues**:
   - Search existing issues
   - Open new issue with:
     - Error message
     - Debug logs
     - Configuration (hide credentials!)
     - Steps to reproduce

4. **Test in Isolation**:
   - Try each action separately
   - Use minimal configuration
   - Test with test account

---

## Emergency Procedures

### If Locked Out of LinkedIn

1. Stop all automation immediately
2. Try logging in manually
3. Complete any verification steps
4. Wait 24-48 hours
5. Contact LinkedIn support if needed

### If Application Won't Stop

```bash
# Find process
ps aux | grep subspace

# Kill process
killall subspace

# Force kill if needed
killall -9 subspace
```

### If Database Corrupted

```bash
# Backup
cp data/subspace.db data/subspace.db.corrupted

# Try repair
sqlite3 data/subspace.db "PRAGMA integrity_check;"

# Last resort: delete and start fresh
rm data/subspace.db
```

---

## Prevention Checklist

Before running:
- [ ] Using test account (not main account)
- [ ] Conservative rate limits configured
- [ ] All stealth features enabled
- [ ] Operating hours set appropriately
- [ ] Realistic delays configured (2-5 seconds)
- [ ] Understand risks and Terms of Service

---

**Remember**: This tool is for educational purposes. LinkedIn automation violates their ToS. Use responsibly and at your own risk.
