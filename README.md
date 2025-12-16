# SubSpace - LinkedIn Automation Tool

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> ⚠️ **CRITICAL DISCLAIMER**: This project is for **EDUCATIONAL PURPOSES ONLY**. Automating LinkedIn violates their Terms of Service and may result in permanent account bans. **DO NOT USE IN PRODUCTION**.

A comprehensive technical proof-of-concept demonstrating advanced browser automation, anti-detection techniques, and clean Go architecture for LinkedIn automation.

## 🎯 Project Overview

SubSpace showcases sophisticated browser automation capabilities with a focus on:
- **Human-like behavior simulation** using Bézier curves and randomized patterns
- **Advanced anti-bot detection** with 10+ stealth techniques
- **Clean, modular architecture** following Go best practices
- **Robust state management** with SQLite persistence
- **Comprehensive logging** and error handling

## ✨ Features

### Core Functionality

- **Authentication System**
  - Secure credential handling via environment variables
  - Session persistence with cookie management
  - Automatic security checkpoint detection (2FA, captcha)
  - Graceful login failure handling

- **Search & Targeting**
  - Search by job title, company, location, or keywords
  - Intelligent pagination handling
  - Duplicate profile detection
  - Search history tracking

- **Connection Requests**
  - Automated connection request sending
  - Personalized note support with template variables
  - Daily and hourly rate limiting
  - Request status tracking

- **Messaging System**
  - Detect newly accepted connections
  - Automated follow-up messages
  - Template support with dynamic variables (`{firstName}`, `{lastName}`, `{company}`)
  - Message history persistence

### 🛡️ Anti-Detection Mechanisms

SubSpace implements **10 sophisticated stealth techniques** to mimic human behavior:

#### Mandatory Techniques (3/3 Implemented)

1. **Human-like Mouse Movement** ✅
   - Bézier curve trajectories with 2 control points
   - Variable speed (slower at start/end, faster in middle)
   - Micro-corrections and occasional overshoot
   - Natural hovering before clicks

2. **Randomized Timing Patterns** ✅
   - Random delays between actions (2-5 seconds default)
   - Human "thinking time" simulation (1-4 seconds)
   - 10% chance of longer pauses (distraction simulation)
   - Variable keystroke intervals

3. **Browser Fingerprint Masking** ✅
   - `navigator.webdriver` removal
   - Random realistic user agents
   - Plugin and language property masking
   - Screen properties randomization
   - Chrome runtime mocking

#### Additional Techniques (7/5 Required)

4. **Random Scrolling Behavior** ✅
   - Variable scroll speeds with acceleration/deceleration
   - Occasional scroll-back movements (20% chance)
   - Smooth scrolling with multiple steps
   - Viewport-aware scrolling patterns

5. **Realistic Typing Simulation** ✅
   - Variable keystroke intervals (150-300ms)
   - 5% chance of typos with corrections
   - Occasional longer pauses (thinking/checking)
   - Backspace pattern simulation

6. **Mouse Hovering & Movement** ✅
   - Random hover events over elements
   - Natural cursor wandering
   - Realistic hover duration (300-1000ms)

7. **Activity Scheduling** ✅
   - Configurable operating hours (9 AM - 5 PM default)
   - Automatic waiting outside operating hours
   - Business hour enforcement

8. **Rate Limiting & Throttling** ✅
   - Daily connection request limits (20/day default)
   - Hourly connection request limits (5/hour default)
   - Daily message limits (15/day default)
   - Automatic quota enforcement

9. **Random Break Simulation** ✅
   - 5% chance of breaks after any action
   - Variable break duration (3-10 minutes)
   - Mimics coffee breaks, distractions

10. **Random Viewport Sizes** ✅
    - 7 realistic viewport configurations
    - Common laptop/desktop resolutions
    - Randomized on each session

## 🏗️ Architecture

```
SubSpace/
├── main.go                 # Application entry point & CLI
├── pkg/
│   ├── auth/              # Authentication & session management
│   │   └── auth.go
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── connection/        # Connection request handling
│   │   └── connection.go
│   ├── messaging/         # Messaging operations
│   │   └── messaging.go
│   ├── search/            # Search & profile discovery
│   │   └── search.go
│   ├── stealth/           # Anti-detection mechanisms
│   │   ├── stealth.go     # Human behavior simulation
│   │   └── browser.go     # Browser fingerprint masking
│   └── storage/           # Database & persistence
│       └── database.go
├── data/                  # Database & session storage
│   ├── subspace.db       # SQLite database
│   └── sessions/         # Session cookies
├── go.mod                 # Go module dependencies
├── go.sum                 # Dependency checksums
├── .env                   # Environment configuration (not in repo)
├── .env.example           # Environment template
├── .gitignore
└── README.md
```

### Package Responsibilities

- **`auth`**: LinkedIn authentication, session persistence, security checkpoint detection
- **`config`**: YAML/environment variable configuration, validation
- **`connection`**: Connection request automation, rate limiting, status tracking
- **`messaging`**: Message sending, template processing, connection detection
- **`search`**: People search, pagination, result extraction
- **`stealth`**: Human behavior simulation, timing randomization, mouse movement
- **`storage`**: SQLite database operations, request/message tracking

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** installed ([Download](https://go.dev/dl/))
- **Google Chrome** or **Chromium** browser
- LinkedIn account credentials

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/saurabhkuntal/subspace.git
   cd subspace
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your LinkedIn credentials
   ```

4. **Build the application**
   ```bash
   go build -o subspace
   ```

### Configuration

Edit `.env` file with your credentials:

```env
# LinkedIn Credentials
LINKEDIN_EMAIL=your.email@example.com
LINKEDIN_PASSWORD=your_password_here

# Application Settings
LOG_LEVEL=info              # debug, info, warn, error
HEADLESS=false              # true for headless mode
SLOW_MOTION=0               # Slow motion in ms (for debugging)

# Connection Limits
MAX_CONNECTIONS_PER_DAY=20
MAX_CONNECTIONS_PER_HOUR=5

# Messaging Settings
MAX_MESSAGES_PER_DAY=15
DEFAULT_MESSAGE_TEMPLATE=Hi {firstName}, I'd love to connect!

# Stealth Configuration
MIN_ACTION_DELAY=2000       # Minimum delay in ms
MAX_ACTION_DELAY=5000       # Maximum delay in ms
ENABLE_RANDOM_SCROLL=true
ENABLE_TYPING_SIMULATION=true
ENABLE_MOUSE_HOVER=true
OPERATING_HOURS_START=9     # 24-hour format
OPERATING_HOURS_END=17

# Database & Session
DATABASE_PATH=./data/subspace.db
SESSION_PATH=./data/sessions
```

## 📖 Usage

### Search for People

```bash
./subspace -action=search \
  -keywords="Software Engineer,Product Manager" \
  -location="San Francisco" \
  -max-results=20
```

### Send Connection Requests

```bash
./subspace -action=connect \
  -keywords="DevOps Engineer" \
  -location="United States" \
  -max-results=10 \
  -message="Hi {firstName}, I noticed we share similar interests in cloud technologies!"
```

### Send Messages to New Connections

```bash
./subspace -action=message \
  -message="Thanks for connecting! Looking forward to staying in touch."
```

### Detect New Connections

```bash
./subspace -action=detect-connections
```

### Command-Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-action` | Action to perform: `search`, `connect`, `message`, `detect-connections` | `search` |
| `-keywords` | Comma-separated search keywords | From config |
| `-location` | Search location filter | From config |
| `-max-results` | Maximum search results to retrieve | `10` |
| `-message` | Custom message for requests/messages | From config |
| `-config` | Path to YAML configuration file | - |
| `-clear-session` | Clear saved session before starting | `false` |

### Template Variables

Use these variables in your messages for personalization:

- `{firstName}` - First name
- `{lastName}` - Last name
- `{name}` - Full name
- `{title}` - Job title
- `{company}` - Company name

## 🗄️ Database Schema

SubSpace uses SQLite for persistence:

### Tables

**connection_requests**
- `id` - Auto-incrementing primary key
- `profile_url` - LinkedIn profile URL (unique)
- `profile_name` - User's full name
- `company` - Company name
- `message` - Custom message sent
- `sent_at` - Timestamp
- `status` - `pending`, `accepted`, `rejected`

**messages**
- `id` - Auto-incrementing primary key
- `profile_url` - LinkedIn profile URL
- `profile_name` - User's full name
- `content` - Message content
- `sent_at` - Timestamp

**search_history**
- `id` - Auto-incrementing primary key
- `query` - Search keywords
- `location` - Location filter
- `result_count` - Number of results
- `searched_at` - Timestamp

## 🧪 Testing & Development

### Run with Debug Logging

```bash
LOG_LEVEL=debug ./subspace -action=search
```

### Run with Visible Browser (Non-Headless)

```bash
HEADLESS=false ./subspace -action=connect
```

### Slow Motion Mode (for Debugging)

```bash
SLOW_MOTION=500 ./subspace -action=search
```

### Clear Session and Re-login

```bash
./subspace -clear-session -action=search
```

## 🔐 Security Considerations

- ✅ Credentials stored in `.env` (gitignored)
- ✅ Session cookies stored locally
- ✅ No hardcoded credentials in code
- ✅ Database stored locally (not exposed)
- ⚠️ Use strong passwords for LinkedIn accounts
- ⚠️ Consider using a test account for development

## 📊 Code Quality

### Implemented Best Practices

- **Modular Architecture**: Separated concerns across packages
- **Error Handling**: Comprehensive error checking with context
- **Logging**: Leveled logging (debug, info, warn, error) with structured output
- **Configuration Management**: Environment variables + YAML support
- **State Persistence**: SQLite for reliable data storage
- **Documentation**: Inline comments, function documentation, README

### Code Statistics

- **Packages**: 7 core packages
- **Functions**: 80+ documented functions
- **Lines of Code**: 2,500+ lines
- **Stealth Techniques**: 10 implemented
- **Test Coverage**: Database operations, configuration loading

## 🚧 Known Limitations

- LinkedIn's UI changes frequently; selectors may need updates
- 2FA/captcha requires manual intervention
- Rate limits are enforced by LinkedIn (not just this tool)
- Headless mode may trigger additional detection
- Some profiles may have restricted connection options

## 🤝 Contributing

This is an educational project. Contributions are welcome for:
- Additional stealth techniques
- Improved error handling
- Better LinkedIn selector stability
- Documentation improvements

## 📄 License

MIT License - see LICENSE file for details

## ⚖️ Legal Disclaimer

**This tool is provided for educational and research purposes only.**

- ❌ **Do NOT use this tool on live LinkedIn accounts**
- ❌ **Do NOT deploy this tool in production environments**
- ❌ **Do NOT use this tool to spam or harass users**
- ⚠️ Using this tool violates [LinkedIn's Terms of Service](https://www.linkedin.com/legal/user-agreement)
- ⚠️ Automated activity may result in permanent account suspension
- ⚠️ The authors assume NO responsibility for misuse

**Use at your own risk. You are solely responsible for any consequences.**

## 📞 Support

For educational inquiries or technical questions:
- Open an issue on GitHub
- Review existing documentation
- Check the demo video (see below)

## 🎥 Demonstration Video

A comprehensive walkthrough video demonstrating SubSpace's features, setup process, and capabilities is available at:

**[Demo Video Link]** *(To be added after recording)*

The video covers:
- Environment setup and configuration
- Authentication and session management
- Search functionality with various filters
- Connection request automation with personalization
- Messaging system and new connection detection
- Stealth mechanism demonstrations
- Database inspection and state tracking

---

**Built with ❤️ for learning and technical demonstration purposes**

*Remember: With great power comes great responsibility. Use automation ethically and respect platforms' terms of service.*
