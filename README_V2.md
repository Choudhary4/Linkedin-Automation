# 🚀 SubSpace LinkedIn Automation - Advanced Anti-Detection Edition

## ⚠️ Important Disclaimer

**This tool is for educational and research purposes only.** Using automated tools on LinkedIn violates their Terms of Service and may result in account restrictions or bans. The techniques implemented here demonstrate advanced web automation and anti-detection methods used in the industry.

**Use at your own risk. We are not responsible for any consequences.**

---

## 🎯 What's New in v2.0

### Advanced Anti-Detection Features

✅ **Browser Fingerprint Spoofing**
- Canvas fingerprinting protection with noise injection
- Audio context fingerprinting evasion
- WebRTC IP leak prevention  
- Font rendering randomization
- Screen property spoofing
- Battery API mocking
- Timezone consistency enforcement

✅ **Residential Proxy Support**
- Smart proxy rotation with success rate tracking
- Geographic targeting (match profile location)
- Automatic failover and health monitoring
- Support for residential, mobile, and datacenter proxies

✅ **Session Building & Account Aging**
- 4-week gradual warm-up schedule
- Natural activity simulation (browsing, liking, reading)
- Reputation scoring system
- Daily action limits with gradual increase
- Human-like usage pattern simulation

✅ **Human Behavior Simulation**
- Extended session warm-up (visits feed first)
- Variable scrolling patterns (20-30 seconds)
- Random delays and pauses
- "Distraction" simulation
- Peak hour activity adjustment

---

## 📦 Installation

```bash
# Clone repository
git clone https://github.com/yourusername/subspace.git
cd subspace

# Install dependencies
go mod download

# Build
go build -o subspace main.go

# Or use the launcher script
chmod +x run.sh
./run.sh
```

---

## 🎮 Quick Start

### 1. Basic Usage (Quick Test)

```bash
# Run interactive mode
./run.sh

# Login manually when browser opens
# Browser stays visible (HEADLESS=false)
# Choose from 8 menu options
```

### 2. Advanced Usage (Recommended)

**Step 1: Configure Proxies** (Optional but highly recommended)

Edit `config.yaml`:
```yaml
proxies:
  - host: "residential-proxy.com"
    port: "8080"
    username: "your_username"
    password: "your_password"
    country: "US"
    city: "New York"
    type: "residential"
```

**Step 2: Account Warm-Up** (Critical for new accounts)

```bash
./run.sh

# Select Option 7: Account warm-up session
# Duration: 30-60 minutes recommended
# Repeat daily for 4 weeks before automation
```

**Step 3: Use Automation Features**

After warm-up period, use search, connect, and message features.

---

## 📋 Feature Comparison

| Feature | Basic Version | Advanced Version |
|---------|---------------|------------------|
| **Browser Stealth** | ✓ Basic flags | ✓✓ Advanced fingerprinting |
| **Proxy Support** | ✗ None | ✓✓ Residential rotation |
| **Session Building** | ✗ None | ✓✓ 4-week warm-up |
| **Human Behavior** | ✓ Basic delays | ✓✓ Advanced patterns |
| **Canvas Protection** | ✗ None | ✓✓ Noise injection |
| **Audio Protection** | ✗ None | ✓✓ Randomization |
| **WebRTC Protection** | ✗ None | ✓✓ IP leak prevention |
| **Detection Risk** | 🔴 High | 🟡 Medium-Low |

---

## 🛡️ Anti-Detection Techniques Explained

### 1. Fingerprint Spoofing

**Problem:** Websites can uniquely identify browsers using canvas/audio/WebGL fingerprints.

**Solution:**
```javascript
// Canvas: Adds imperceptible noise to pixel data
canvas.toDataURL() → Modified with ±1-2 RGB variation

// Audio: Randomizes audio processing signature
audioContext.getChannelData() → Noise added

// WebRTC: Blocks IP-leaking ICE candidates
RTCPeerConnection → Only relay candidates allowed
```

### 2. Proxy Rotation

**Problem:** LinkedIn tracks IPs and flags datacenter/repeated IPs.

**Solution:**
- Residential proxies from real ISPs (Verizon, AT&T, Airtel)
- Geographic matching (Pune profile → Pune proxy)
- Smart rotation based on success rates
- 5-minute minimum between reuses

### 3. Session Building

**Problem:** New accounts with immediate high activity are flagged.

**Solution:**
```
Week 1: Browse only (15-30 min/day)
Week 2: Light engagement (30-40 min/day, 1-2 connections)
Week 3: Moderate activity (40-45 min/day, 3-5 connections)
Week 4+: Full automation (50 actions/day max)
```

### 4. Human Patterns

**Problem:** Bots have consistent, inhuman timing patterns.

**Solution:**
- Variable delays (1-9 seconds)
- Random scroll amounts (250-450px)
- Occasional backscroll (mimics reading)
- Peak hour adjustment (12pm, 6pm more active)
- Weekend reduction
- Random "distraction" pauses (3-5s)

---

## 📊 Effectiveness Analysis

### Detection Rates (Estimated)

| Configuration | Detection Rate | Account Ban Risk |
|---------------|----------------|------------------|
| No Protection | ~95% | Very High |
| Basic Stealth | ~60% | High |
| Advanced Stealth (no warm-up) | ~35% | Medium |
| **Full Advanced (with warm-up)** | **~15-20%** | **Low-Medium** |

### Success Rate by Account Age

| Account Age | Recommended Actions/Day | Ban Risk |
|-------------|-------------------------|----------|
| < 1 week | 5 | High |
| 1-2 weeks | 10-15 | Medium-High |
| 2-4 weeks | 20-30 | Medium |
| > 4 weeks | 40-50 | Low-Medium |
| > 3 months | 50-75 | Low |

---

## 🔧 Configuration

### Complete Config Example

```yaml
# config.yaml

linkedin:
  email: "your_email@example.com"
  password: "your_password"

app:
  log_level: "info"

stealth:
  min_action_delay: 2000
  max_action_delay: 5000
  enable_random_scroll: true
  enable_typing_simulation: true
  enable_mouse_hover: true
  operating_hours_start: 8
  operating_hours_end: 22

advanced_stealth:
  enable_fingerprinting: true
  enable_proxy_rotation: true
  enable_session_building: true
  canvas_noise_level: 0.1
  audio_noise_level: 0.00005

proxies:
  - host: "proxy1.example.com"
    port: "8080"
    username: "user1"
    password: "pass1"
    country: "US"
    city: "New York"
    isp: "Verizon"
    type: "residential"
    
  - host: "proxy2.example.com"
    port: "8080"
    country: "India"
    city: "Pune"
    isp: "Airtel"
    type: "residential"

search:
  default_keywords:
    - "Software Engineer"
    - "Product Manager"
  default_location: "United States"

connections:
  max_per_day: 50
  max_per_hour: 10

messaging:
  max_per_day: 30
  default_message_template: "Thanks for connecting!"

database:
  path: "./data/subspace.db"

session:
  path: "./data/session.json"
```

---

## 📚 Documentation

- **[ANTI_DETECTION.md](ANTI_DETECTION.md)** - Basic anti-detection techniques
- **[ADVANCED_ANTI_DETECTION.md](ADVANCED_ANTI_DETECTION.md)** - Advanced features guide
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Development setup
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues

---

## 🎯 Menu Options

### Interactive Mode (./run.sh)

```
1. 🔍 Search for people
   - Find profiles by keywords and location
   - Enhanced with 20-30 second warm-up per search

2. 🤝 Search and send connection requests
   - Search + automated connection requests
   - Respects daily limits

3. 💬 Send messages to new connections
   - Detects new connections
   - Sends personalized messages

4. 👥 View new connections
   - Lists recent connections

5. 📊 View activity statistics
   - Shows automation history
   - Success rates

6. ⚙️ Change settings
   - View current configuration

7. 🔥 Account warm-up session ← NEW!
   - Simulates natural browsing
   - Recommended: 30-60 min daily for 4 weeks

8. 🚪 Exit
   - Safely close browser and exit
```

---

## 🚨 Current Limitations

### What Still Gets Detected

Even with all protections, LinkedIn may still detect automation due to:

1. **Server-Side Patterns**
   - Request timing analysis
   - Behavioral ML models
   - Historical account analysis

2. **Infrastructure Detection**
   - Proxy IP reputation
   - Datacenter IP ranges
   - Known automation signatures

3. **Human Review**
   - Manual reports from users
   - Suspicious activity flags
   - Pattern anomalies

### Known Issues

❌ **Search Results:** LinkedIn currently serves skeleton placeholders even with all protections. This is due to:
- Sophisticated ML-based bot detection
- Server-side behavioral analysis
- Account reputation requirements

✅ **What Works:**
- Manual login with session persistence
- Browser fingerprint spoofing
- Proxy infrastructure
- Session building framework
- Connection/message features (when profile URLs provided manually)

---

## 🔬 Testing Your Setup

### Test Fingerprint Protection

Visit these sites to verify protections:
- https://pixelscan.net/ (comprehensive test)
- https://abrahamjuliot.github.io/creepjs/ (detailed fingerprint)
- https://browserleaks.com/ (leak detection)
- https://whoer.net/ (anonymity score)

**Expected Results:**
- Canvas: Unique per session, no detection
- Audio: Randomized signature
- WebRTC: No IP leaks
- Proxy: Location matches proxy

### Test Proxy Rotation

```bash
# Check current IP
curl https://api.ipify.org

# Through proxy
curl --proxy http://user:pass@proxy:port https://api.ipify.org

# Should show different IPs
```

---

## 🎓 Learning Resources

### Understand the Techniques

**Research Papers:**
- "Fingerprinting the Fingerprinters" - Browser fingerprint detection
- "Canvas Fingerprinting" - Pixel-based tracking
- "Audio Context Fingerprinting" - Sound-based identification

**Tools to Study:**
- [Puppeteer Extra Stealth](https://github.com/berstend/puppeteer-extra) - Similar techniques for Node.js
- [FingerprintJS](https://github.com/fingerprintjs/fingerprintjs) - Fingerprinting library
- [CreepJS](https://abrahamjuliot.github.io/creepjs/) - Detection testing

---

## 🛠️ Troubleshooting

### Still Getting Detected?

**1. Complete Full Warm-Up**
```bash
# Run warm-up daily for 28 days minimum
./run.sh → Option 7 → 30-60 minutes
```

**2. Verify Proxy Type**
```bash
# Must be residential, not datacenter
# Check with: https://whatismyipaddress.com/
```

**3. Reduce Action Frequency**
```yaml
# config.yaml
connections:
  max_per_day: 20  # Start lower
  max_per_hour: 5
```

**4. Check Account History**
- New accounts: Higher risk
- Previously restricted: Permanently flagged
- Aged accounts (3+ months): Lower risk

### No Search Results?

This is expected! LinkedIn's detection is very sophisticated. Options:

1. **Use Manual Search + Automation**
   - Search manually on LinkedIn
   - Copy profile URLs
   - Use tool for connection/messaging

2. **Focus on Other Features**
   - Connection management works
   - Messaging existing connections works
   - Network viewing works

3. **Wait for Account Aging**
   - Continue warm-up for 8-12 weeks
   - Gradually increase activity
   - Build genuine engagement history

---

## ⚖️ Legal & Ethical Considerations

### Terms of Service

Using this tool violates LinkedIn's Terms of Service section 8.2:
> "Don't develop, support or use software, devices, scripts, robots or any other means or processes... to scrape the Services or otherwise copy profiles and other data from the Services."

### Recommended Use

This tool is intended for:
- ✅ Learning web automation techniques
- ✅ Understanding anti-detection methods
- ✅ Security research
- ✅ Educational demonstrations

**NOT for:**
- ❌ Commercial lead generation at scale
- ❌ Spam or harassment
- ❌ Violation of user privacy
- ❌ Terms of Service violations

### Consequences

Possible outcomes of automation:
- Account restrictions
- Permanent account ban
- IP address blocking
- Legal action (in extreme cases)

---

## 🤝 Contributing

Contributions welcome! Areas of interest:
- Additional fingerprinting protections
- Better proxy management
- Machine learning behavior modeling
- Official API integration
- Testing and documentation

---

## 📄 License

MIT License - See LICENSE file for details.

**Disclaimer:** This tool is provided "as is" without warranty. Use at your own risk. The authors are not responsible for any consequences resulting from use of this tool.

---

## 🙏 Acknowledgments

Anti-detection techniques inspired by:
- Puppeteer Extra Stealth Plugin
- FingerprintJS research
- CreepJS detection methods
- Research papers on browser fingerprinting

---

## 📞 Support

- **Issues:** [GitHub Issues](https://github.com/yourusername/subspace/issues)
- **Discussions:** [GitHub Discussions](https://github.com/yourusername/subspace/discussions)
- **Documentation:** See `/docs` folder

---

**Version:** 2.0.0  
**Last Updated:** December 2025  
**Status:** ⚠️ Educational/Research - LinkedIn search currently blocked by detection

