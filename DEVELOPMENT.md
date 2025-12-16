# SubSpace Development Guide

## Project Structure

### Core Packages

#### `pkg/auth`
Authentication and session management. Handles login, session persistence, and security checkpoint detection.

**Key Functions:**
- `Login(page)` - Performs LinkedIn login with human-like behavior
- `SaveSession(page)` - Persists session cookies to disk
- `LoadSession(page)` - Loads saved session
- `IsLoggedIn(page)` - Validates current session
- `HasSecurityCheckpoint(page)` - Detects 2FA/captcha

#### `pkg/config`
Configuration management with environment variable and YAML support.

**Key Functions:**
- `LoadConfig(yamlPath)` - Loads configuration from file/env
- `Validate()` - Validates configuration values

#### `pkg/stealth`
Anti-detection mechanisms and human behavior simulation.

**Key Functions:**
- `HumanMouseMove(page, element)` - Bezier curve mouse movement
- `HumanClick(page, element)` - Realistic click simulation
- `HumanType(page, element, text)` - Typed text with variations
- `RandomScroll(page)` - Natural scrolling behavior
- `RandomDelay()` - Variable timing between actions
- `IsWithinOperatingHours()` - Operating hours enforcement

#### `pkg/search`
LinkedIn people search and profile discovery.

**Key Functions:**
- `SearchPeople(page, keywords, location, maxResults)` - Main search
- `SearchByJobTitle(page, jobTitle, location, maxResults)` - Title-based
- `SearchByCompany(page, company, location, maxResults)` - Company-based
- `extractResultsFromPage(page)` - Parse search results

#### `pkg/connection`
Connection request automation with rate limiting.

**Key Functions:**
- `SendConnectionRequest(page, profile, message)` - Send single request
- `SendBulkConnectionRequests(page, profiles, message)` - Bulk sending
- `GetPendingConnections()` - Retrieve pending requests
- `UpdateConnectionStatus(profileURL, status)` - Update status

#### `pkg/messaging`
Message sending and connection detection.

**Key Functions:**
- `SendMessage(page, profileURL, name, message)` - Send message
- `SendBulkMessages(page, recipients)` - Bulk messaging
- `DetectNewConnections(page)` - Find accepted connections
- `PersonalizeMessage(template, profile)` - Template processing

#### `pkg/storage`
SQLite database operations for persistence.

**Key Functions:**
- `SaveConnectionRequest(request)` - Save request record
- `GetConnectionRequestCountToday()` - Check daily count
- `HasSentConnectionRequest(profileURL)` - Duplicate check
- `SaveMessage(message)` - Save message record
- `SaveSearchHistory(search)` - Save search record

## Adding New Features

### Example: Adding a New Stealth Technique

1. **Add to `pkg/stealth/stealth.go`:**

```go
// RandomMouseJiggle adds small random mouse movements
func (m *Manager) RandomMouseJiggle(page *rod.Page) error {
    // Get current position
    currentX, currentY := 100.0, 100.0
    
    // Small random movements
    for i := 0; i < 3; i++ {
        offsetX := (m.rand.Float64() - 0.5) * 10
        offsetY := (m.rand.Float64() - 0.5) * 10
        
        err := page.Mouse.Move(currentX+offsetX, currentY+offsetY, 1)
        if err != nil {
            return err
        }
        
        time.Sleep(time.Duration(50+m.rand.Intn(100)) * time.Millisecond)
    }
    
    return nil
}
```

2. **Update configuration in `pkg/config/config.go`:**

```go
type StealthConfig struct {
    // ... existing fields
    EnableMouseJiggle bool `yaml:"enable_mouse_jiggle"`
}
```

3. **Add to `.env.example`:**

```env
ENABLE_MOUSE_JIGGLE=true
```

4. **Use in automation flow:**

```go
// In connection/connection.go
if err := m.stealthMgr.RandomMouseJiggle(page); err != nil {
    m.logger.Warnf("Failed to jiggle mouse: %v", err)
}
```

## Testing

### Manual Testing

```bash
# Test with debug logging
LOG_LEVEL=debug ./subspace -action=search -keywords="test"

# Test with visible browser
HEADLESS=false ./subspace -action=search

# Test with slow motion
SLOW_MOTION=1000 ./subspace -action=search
```

### Database Inspection

```bash
# Open database
sqlite3 data/subspace.db

# View connection requests
SELECT * FROM connection_requests ORDER BY sent_at DESC LIMIT 10;

# View messages
SELECT * FROM messages ORDER BY sent_at DESC LIMIT 10;

# Check daily counts
SELECT COUNT(*) FROM connection_requests 
WHERE DATE(sent_at) = DATE('now');
```

## Common Issues & Solutions

### Issue: "Failed to find element"
**Solution:** LinkedIn's UI changes frequently. Update selectors in respective packages.

```go
// Try multiple selectors
selectors := []string{
    "button[aria-label*='Connect']",
    "button:has-text('Connect')",
    ".pvs-profile-actions button:has-text('Connect')",
}
```

### Issue: "Login failed"
**Solution:** 
1. Check credentials in `.env`
2. Clear session: `./subspace -clear-session`
3. Try with `HEADLESS=false` to see what's happening

### Issue: "Reached daily limit"
**Solution:** Limits are enforced by database. Either wait or manually clear:

```sql
DELETE FROM connection_requests WHERE sent_at < DATE('now');
```

## Performance Optimization

### Reduce Memory Usage

```go
// Close pages when done
defer page.Close()

// Limit result size
maxResults := 10  // Don't fetch thousands
```

### Speed Up Testing

```go
// Disable stealth features during development
cfg.Stealth.EnableRandomScroll = false
cfg.Stealth.EnableTypingSimulation = false
cfg.Stealth.MinActionDelay = 100
cfg.Stealth.MaxActionDelay = 500
```

## Security Best Practices

1. **Never commit `.env` file**
2. **Use test accounts for development**
3. **Keep dependencies updated:** `go get -u ./...`
4. **Review logs for sensitive data leaks**
5. **Use environment-specific configs**

## Deployment Checklist

- [ ] Remove or obfuscate demo credentials
- [ ] Test with fresh `.env.example`
- [ ] Verify `.gitignore` excludes sensitive files
- [ ] Update `README.md` with any new features
- [ ] Test build process: `go build -o subspace`
- [ ] Verify all dependencies are in `go.mod`
- [ ] Add disclaimer to all public-facing docs
- [ ] Test on clean machine/VM

## Contributing Guidelines

1. **Fork the repository**
2. **Create feature branch:** `git checkout -b feature/new-stealth-technique`
3. **Follow existing code style**
4. **Add inline documentation**
5. **Test thoroughly**
6. **Submit pull request with description**

## Additional Resources

- [Rod Documentation](https://go-rod.github.io/)
- [Go Best Practices](https://golang.org/doc/effective_go)
- [LinkedIn Terms of Service](https://www.linkedin.com/legal/user-agreement)
- [Browser Automation Patterns](https://www.selenium.dev/documentation/)

---

**Happy Coding! Remember: Use responsibly and ethically.**
