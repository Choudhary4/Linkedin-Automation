# Anti-Detection Techniques Implemented

This document outlines all the anti-detection and stealth techniques implemented in SubSpace to reduce the risk of being flagged by LinkedIn's automation detection systems.

## 🛡️ Browser-Level Anti-Detection

### Browser Launch Flags
Located in: `pkg/stealth/browser.go`

```
- disable-blink-features: AutomationControlled
- exclude-switches: enable-automation  
- disable-infobars
- disable-web-security
- disable-features: IsolateOrigins,site-per-process
- allow-running-insecure-content
- disable-ipc-flooding-protection
- window-size: 1920,1080
- Custom User-Agent: Chrome 120 on macOS
```

### JavaScript Fingerprint Masking
Located in: `pkg/stealth/browser.go - ApplyStealthSettings()`

**Navigator Properties:**
- Overrides `navigator.webdriver` to return `undefined`
- Mocks realistic browser plugins (PDF viewers, Native Client)
- Sets proper `navigator.languages` to `['en-US', 'en']`
- Overrides `navigator.userAgentData` with realistic Chrome/Chromium brands
- Removes `navigator.__proto__.webdriver`

**Chrome Runtime:**
- Creates realistic `window.chrome` object with `runtime`, `loadTimes`, `csi`, `app`

**WebGL Fingerprint:**
- Overrides `WebGLRenderingContext.getParameter` to return:
  - Vendor: "Intel Inc."
  - Renderer: "Intel Iris OpenGL Engine"

**Permissions API:**
- Intercepts `navigator.permissions.query` for notifications

**Proxy Detection Hiding:**
- Wraps toString methods to hide Proxy usage in critical properties

## 🧑‍💻 Human Behavior Simulation

### Session Warm-Up
Located in: `pkg/search/search.go - SearchPeople()`

Before performing searches:
1. Visits LinkedIn homepage (`/feed/`)
2. Waits 3-6 seconds (randomized)
3. Performs 2 random scrolls (200-400px each)
4. Random pauses between actions (1-2 seconds)

### Realistic Scrolling Patterns
Located in: `pkg/search/search.go - extractResultsFromPage()`

**Initial Page Load:**
- Wait 6-9 seconds (randomized)
- Simulate reading with 4 pauses (600-1200ms each)

**Scrolling Behavior:**
- 7 scroll iterations with varying amounts (250-450px)
- Variable scroll speeds (6-12 steps)
- Random pauses between scrolls (2-4 seconds)
- Occasional scroll-ups to mimic reading (at positions 2 and 5)
- Long pause simulation at position 3 (3-5 seconds - distraction)
- Total scroll time: ~20-30 seconds

### Random Delays
Located in: `pkg/stealth/stealth.go`

**RandomDelay():**
- Configured min/max delay from config
- Applied between major actions

**HumanDelay():**
- Base: 1-4 seconds
- 10% chance of longer pause (4-9 seconds total)
- Simulates human thinking/distraction

## 🔍 LinkedIn-Specific Protections

### Mobile Redirect Detection
Located in: `pkg/search/search.go - SearchPeople()`

- Detects `linkedin.com/m/` URLs
- Automatically redirects to desktop version
- Waits 3 seconds after redirect

### Security Check Detection
Located in: `pkg/search/search.go - extractResultsFromPage()`

Checks HTML for:
- "we've detected unusual activity"
- "please verify"
- "security verification"

Returns error if detected, allowing manual intervention.

### Skeleton/Placeholder Detection
Located in: `pkg/search/search.go - extractResultsFromPage()`

**Problem:** LinkedIn serves placeholder elements with `data-chameleon-result-urn="urn:li:member:headless"`

**Solution:**
1. Filters out all elements containing "headless" in URN attribute
2. Falls back to link extraction method if no real profiles found
3. Logs detailed page state information
4. Takes screenshots for debugging

## 📸 Debugging Features

### Screenshot Capture
Located in: `pkg/search/search.go - extractResultsFromPage()`

- Saves timestamped screenshots to `screenshots/` directory
- Helps analyze what LinkedIn is actually showing
- Path format: `screenshots/search_<timestamp>.png`

### Page State Analysis
Located in: `pkg/search/search.go - extractResultsFromPage()`

Runs JavaScript to check:
- Number of skeleton elements
- Number of real profile links  
- Presence of loading indicators

## ⚙️ Configuration

All stealth settings can be configured via `config.yaml`:

```yaml
stealth:
  min_action_delay: 1000        # Min milliseconds between actions
  max_action_delay: 3000        # Max milliseconds between actions
  enable_random_scroll: true    # Random scrolling
  enable_typing_simulation: true # Human-like typing
  enable_mouse_hover: true      # Mouse hover simulation
  operating_hours_start: 8      # Start hour (24h format)
  operating_hours_end: 22       # End hour (24h format)

browser:
  headless: false               # MUST be false for LinkedIn
  slow_motion: 0                # Milliseconds between actions
```

## 📊 Current Status

### ✅ Working:
- Manual login with session persistence
- Browser stealth settings applied
- Human behavior simulation active
- Mobile redirect handling
- Security check detection
- Skeleton element filtering

### ⚠️ Limitations:
LinkedIn's anti-bot detection still identifies automation and serves skeleton content. This is likely due to:

1. **Traffic Analysis:** LinkedIn analyzes patterns across sessions
2. **Machine Learning Models:** Sophisticated behavioral analysis
3. **Browser Fingerprinting:** Even with masking, some signals remain
4. **Account History:** New/suspicious accounts flagged more aggressively

### 🔮 Potential Improvements (Advanced):

1. **Residential Proxies:** Rotate IP addresses from real residential locations
2. **Longer Session Building:** Use account normally for days before automation
3. **Puppeteer Extra Stealth:** Port techniques from puppeteer-extra-plugin-stealth
4. **Canvas Fingerprinting:** Randomize canvas/font rendering
5. **Audio Context:** Mock AudioContext fingerprinting
6. **Battery API:** Mock battery status
7. **WebRTC:** Control WebRTC IP leaks
8. **Official API:** Use LinkedIn's official API (requires business approval)

## 🎓 Educational Purpose

**Important:** This tool is for educational purposes to understand web automation and anti-detection techniques. Using it may violate LinkedIn's Terms of Service. The techniques implemented represent common approaches to stealth automation but are not guaranteed to avoid detection.

## 📚 References

- [Puppeteer Stealth Plugin](https://github.com/berstend/puppeteer-extra/tree/master/packages/puppeteer-extra-plugin-stealth)
- [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)
- [Web Automation Detection Methods](https://antoinevastel.com/bot%20detection/2019/07/19/detecting-chrome-headless-v3.html)
- [Go Rod Documentation](https://go-rod.github.io/)
