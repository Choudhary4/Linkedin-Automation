# SubSpace - Quick Reference Guide

## 🚀 Common Commands

### Setup & Build
```bash
# Initial setup
./setup.sh

# Build only
make build
# OR
go build -o subspace main.go

# Install dependencies
go mod download
```

### Search Operations
```bash
# Basic search
./subspace -action=search

# Search with keywords
./subspace -action=search -keywords="Software Engineer" -max-results=10

# Search by location
./subspace -action=search -keywords="Product Manager" -location="San Francisco"

# Search multiple keywords
./subspace -action=search -keywords="DevOps,SRE,Cloud Engineer" -max-results=20
```

### Connection Requests
```bash
# Send connections with default message
./subspace -action=connect -keywords="Software Engineer" -max-results=5

# Send connections with custom message
./subspace -action=connect \
  -keywords="Data Scientist" \
  -max-results=10 \
  -message="Hi {firstName}, I'm impressed by your work in AI!"

# Using template variables
./subspace -action=connect \
  -message="Hi {firstName}, I noticed you work at {company}. Would love to connect!"
```

### Messaging
```bash
# Send messages to new connections
./subspace -action=message -message="Thanks for connecting!"

# With template variables
./subspace -action=message -message="Hi {firstName}, great to connect!"

# Detect new connections first
./subspace -action=detect-connections
```

### Configuration
```bash
# Use custom config file
./subspace -config=config.yaml -action=search

# Clear saved session
./subspace -clear-session -action=search

# With environment variables
LINKEDIN_EMAIL=user@example.com LINKEDIN_PASSWORD=pass ./subspace -action=search
```

---

## 🎛️ Configuration Quick Reference

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LINKEDIN_EMAIL` | - | Your LinkedIn email (required) |
| `LINKEDIN_PASSWORD` | - | Your LinkedIn password (required) |
| `LOG_LEVEL` | `info` | Logging level: debug, info, warn, error |
| `HEADLESS` | `false` | Run browser in headless mode |
| `SLOW_MOTION` | `0` | Slow motion in ms (debugging) |
| `MAX_CONNECTIONS_PER_DAY` | `20` | Daily connection limit |
| `MAX_CONNECTIONS_PER_HOUR` | `5` | Hourly connection limit |
| `MAX_MESSAGES_PER_DAY` | `15` | Daily message limit |
| `MIN_ACTION_DELAY` | `2000` | Min delay between actions (ms) |
| `MAX_ACTION_DELAY` | `5000` | Max delay between actions (ms) |
| `ENABLE_RANDOM_SCROLL` | `true` | Enable scrolling behavior |
| `ENABLE_TYPING_SIMULATION` | `true` | Enable realistic typing |
| `ENABLE_MOUSE_HOVER` | `true` | Enable mouse hovering |
| `OPERATING_HOURS_START` | `9` | Start hour (24h format) |
| `OPERATING_HOURS_END` | `17` | End hour (24h format) |
| `DATABASE_PATH` | `./data/subspace.db` | Database file path |
| `SESSION_PATH` | `./data/sessions` | Session storage path |

### Template Variables

Use in messages and connection notes:

| Variable | Description | Example |
|----------|-------------|---------|
| `{firstName}` | First name | John |
| `{lastName}` | Last name | Doe |
| `{name}` | Full name | John Doe |
| `{title}` | Job title | Software Engineer |
| `{company}` | Company name | Google |

**Example**: `"Hi {firstName}, I see you're a {title} at {company}!"` 
→ `"Hi John, I see you're a Software Engineer at Google!"`

---

## 📊 Database Queries

### Connection Requests
```bash
# Open database
sqlite3 data/subspace.db

# View recent requests
SELECT profile_name, company, sent_at, status 
FROM connection_requests 
ORDER BY sent_at DESC LIMIT 10;

# Count today's requests
SELECT COUNT(*) FROM connection_requests 
WHERE DATE(sent_at) = DATE('now');

# Count this hour's requests
SELECT COUNT(*) FROM connection_requests 
WHERE sent_at >= datetime('now', '-1 hour');

# View pending requests
SELECT profile_name, profile_url, sent_at 
FROM connection_requests 
WHERE status='pending';

# Check if sent to specific person
SELECT * FROM connection_requests 
WHERE profile_url LIKE '%username%';
```

### Messages
```bash
# View recent messages
SELECT profile_name, sent_at, content 
FROM messages 
ORDER BY sent_at DESC LIMIT 10;

# Count today's messages
SELECT COUNT(*) FROM messages 
WHERE DATE(sent_at) = DATE('now');

# Messages to specific person
SELECT * FROM messages 
WHERE profile_url LIKE '%username%';
```

### Search History
```bash
# View recent searches
SELECT query, location, result_count, searched_at 
FROM search_history 
ORDER BY searched_at DESC LIMIT 10;
```

---

## 🔧 Debugging Commands

### Basic Debugging
```bash
# Debug logging
LOG_LEVEL=debug ./subspace -action=search

# Visible browser
HEADLESS=false ./subspace -action=search

# Slow motion (see what's happening)
SLOW_MOTION=1000 HEADLESS=false ./subspace -action=search

# All debug features
LOG_LEVEL=debug HEADLESS=false SLOW_MOTION=500 ./subspace -action=search
```

### Session Management
```bash
# Clear and re-login
./subspace -clear-session -action=search

# Check session files
ls -la data/sessions/

# Remove session manually
rm -f data/sessions/cookies.json
```

### Database Operations
```bash
# Check database size
du -h data/subspace.db

# Backup database
cp data/subspace.db data/subspace.db.backup.$(date +%Y%m%d)

# Reset database (CAUTION!)
rm data/subspace.db

# Check database integrity
sqlite3 data/subspace.db "PRAGMA integrity_check;"
```

---

## 🎨 Customization Examples

### Conservative Settings (Low Detection Risk)
```env
MAX_CONNECTIONS_PER_DAY=10
MAX_CONNECTIONS_PER_HOUR=2
MAX_MESSAGES_PER_DAY=5
MIN_ACTION_DELAY=5000
MAX_ACTION_DELAY=10000
ENABLE_RANDOM_SCROLL=true
ENABLE_TYPING_SIMULATION=true
ENABLE_MOUSE_HOVER=true
OPERATING_HOURS_START=9
OPERATING_HOURS_END=17
```

### Aggressive Settings (Higher Detection Risk)
```env
MAX_CONNECTIONS_PER_DAY=50
MAX_CONNECTIONS_PER_HOUR=10
MAX_MESSAGES_PER_DAY=30
MIN_ACTION_DELAY=1000
MAX_ACTION_DELAY=2000
ENABLE_RANDOM_SCROLL=false
ENABLE_TYPING_SIMULATION=false
ENABLE_MOUSE_HOVER=false
OPERATING_HOURS_START=0
OPERATING_HOURS_END=23
```

### Testing Settings (Fast Development)
```env
MAX_CONNECTIONS_PER_DAY=1000
MAX_CONNECTIONS_PER_HOUR=100
MAX_MESSAGES_PER_DAY=1000
MIN_ACTION_DELAY=100
MAX_ACTION_DELAY=500
HEADLESS=false
SLOW_MOTION=500
LOG_LEVEL=debug
```

---

## 📝 Common Workflows

### Workflow 1: Find & Connect
```bash
# 1. Search for people
./subspace -action=search -keywords="React Developer" -max-results=20

# 2. Review results in logs

# 3. Send connection requests
./subspace -action=connect \
  -keywords="React Developer" \
  -max-results=10 \
  -message="Hi {firstName}, I'm building a React team!"

# 4. Check database
sqlite3 data/subspace.db "SELECT COUNT(*) FROM connection_requests WHERE DATE(sent_at) = DATE('now');"
```

### Workflow 2: Follow-up Messages
```bash
# 1. Detect new connections
./subspace -action=detect-connections

# 2. Send follow-up messages
./subspace -action=message \
  -message="Thanks for connecting, {firstName}! Looking forward to learning more about your work at {company}."

# 3. Verify in database
sqlite3 data/subspace.db "SELECT * FROM messages ORDER BY sent_at DESC LIMIT 5;"
```

### Workflow 3: Daily Automation
```bash
#!/bin/bash
# daily-automation.sh

# Search and connect
./subspace -action=connect \
  -keywords="Software Engineer,Product Manager" \
  -max-results=15 \
  -message="Hi {firstName}, I'd love to connect!"

# Send messages to new connections
./subspace -action=message \
  -message="Thanks for connecting, {firstName}!"

# Check stats
echo "Today's Stats:"
sqlite3 data/subspace.db "SELECT COUNT(*) FROM connection_requests WHERE DATE(sent_at) = DATE('now');" 
sqlite3 data/subspace.db "SELECT COUNT(*) FROM messages WHERE DATE(sent_at) = DATE('now');"
```

---

## 🛡️ Safety Reminders

### Before Running
- [ ] Using test account (not main)
- [ ] Conservative limits set
- [ ] All stealth features enabled
- [ ] Operating hours configured
- [ ] Understand risks

### During Operation
- Monitor logs for errors
- Check for security checkpoints
- Verify rate limits are working
- Watch for unusual behavior

### After Running
- Review sent requests in database
- Check LinkedIn for restrictions
- Back up database
- Clear session if switching accounts

---

## 🔗 Quick Links

- **Repository**: https://github.com/saurabhkuntal/subspace
- **Submission Form**: https://forms.gle/fgbMxgUS19QRKGPa9
- **LinkedIn ToS**: https://www.linkedin.com/legal/user-agreement
- **Go Installation**: https://go.dev/dl/
- **Rod Documentation**: https://go-rod.github.io/

---

## 📞 Getting Help

1. Check [README.md](README.md) for overview
2. Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for issues
3. Check [DEVELOPMENT.md](DEVELOPMENT.md) for technical details
4. Enable debug logging: `LOG_LEVEL=debug`
5. Run with visible browser: `HEADLESS=false`
6. Open GitHub issue with details

---

## ⚡ Performance Tips

### Speed Up Testing
```bash
# Disable stealth features
ENABLE_RANDOM_SCROLL=false \
ENABLE_TYPING_SIMULATION=false \
MIN_ACTION_DELAY=100 \
MAX_ACTION_DELAY=500 \
./subspace -action=search
```

### Reduce Memory
```bash
# Limit results
./subspace -action=search -max-results=5

# Use headless mode
HEADLESS=true ./subspace -action=search
```

### Better Logging
```bash
# Save logs to file
./subspace -action=search 2>&1 | tee subspace.log

# Watch logs in real-time
tail -f subspace.log
```

---

**Remember**: This tool is for educational purposes only. Use responsibly! 🎓
