package auth

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// Authenticator handles LinkedIn authentication
type Authenticator struct {
	email       string
	password    string
	sessionPath string
	logger      *logrus.Logger
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(email, password, sessionPath string, logger *logrus.Logger) *Authenticator {
	rand.Seed(time.Now().UnixNano())

	return &Authenticator{
		email:       email,
		password:    password,
		sessionPath: sessionPath,
		logger:      logger,
	}
}

// ==================== PUBLIC ENTRY ====================

func (a *Authenticator) Login(page *rod.Page) error {
	a.logger.Info("Starting login process...")

	// Try loading existing session
	if err := a.LoadSession(page); err == nil {
		a.logger.Info("Loaded existing session")

		if a.IsLoggedIn(page) {
			a.logger.Info("Session valid, skipping login")
			return nil
		}

		a.logger.Warn("Session expired, clearing cookies")
		_ = a.ClearSession()
	}

	// Fresh login
	if err := a.performFreshLogin(page); err != nil {
		return err
	}

	// Save session
	if err := a.SaveSession(page); err != nil {
		a.logger.Warnf("Failed to save session: %v", err)
	}

	return nil
}

// ManualLogin waits for the user to manually log in to LinkedIn
func (a *Authenticator) ManualLogin(page *rod.Page) error {
	a.logger.Info("Starting manual login process...")

	// Try loading existing session
	if err := a.LoadSession(page); err == nil {
		a.logger.Info("Loaded existing session")

		if a.IsLoggedIn(page) {
			a.logger.Info("Session valid, already logged in")
			return nil
		}

		a.logger.Warn("Session expired, clearing cookies")
		_ = a.ClearSession()
	}

	// Navigate to LinkedIn login page
	a.logger.Info("Navigating to LinkedIn login page...")
	if err := page.Navigate("https://www.linkedin.com/login"); err != nil {
		return fmt.Errorf("failed to navigate to login page: %v", err)
	}

	page.MustWaitLoad()
	time.Sleep(2 * time.Second)

	a.logger.Info("⏳ Please log in manually in the browser window...")
	a.logger.Info("⏳ Waiting for you to complete the login (including any 2FA/verification)...")

	// Wait for user to log in (check every 3 seconds for up to 5 minutes)
	maxWaitTime := 5 * time.Minute
	checkInterval := 3 * time.Second
	elapsed := time.Duration(0)

	for elapsed < maxWaitTime {
		time.Sleep(checkInterval)
		elapsed += checkInterval

		if a.IsLoggedIn(page) {
			a.logger.Info("✓ Manual login successful!")

			// Save the session
			if err := a.SaveSession(page); err != nil {
				a.logger.Warnf("Failed to save session: %v", err)
			} else {
				a.logger.Info("Session saved successfully")
			}

			return nil
		}

		// Show progress every 30 seconds
		if int(elapsed.Seconds())%30 == 0 {
			a.logger.Infof("Still waiting... (%v elapsed)", elapsed)
		}
	}

	return fmt.Errorf("manual login timeout after %v", maxWaitTime)
}

// ==================== CORE LOGIN ====================

func (a *Authenticator) performFreshLogin(page *rod.Page) error {
	a.logger.Info("Performing fresh login")

	if err := page.Navigate("https://www.linkedin.com/login"); err != nil {
		return err
	}

	page.MustWaitLoad()
	page.MustWaitIdle()
	time.Sleep(5 * time.Second)

	if a.HasSecurityCheckpoint(page) {
		return fmt.Errorf("security checkpoint detected before login")
	}

	// Human-like behavior
	a.randomScroll(page)

	// Email
	emailInput, err := a.findEmailInput(page)
	if err != nil {
		return err
	}
	a.humanType(page, emailInput, a.email)

	// Password
	passwordInput, err := a.findPasswordInput(page)
	if err != nil {
		return err
	}
	a.humanType(page, passwordInput, a.password)

	// Submit
	loginBtn, err := page.Timeout(10 * time.Second).Element("button[type='submit']")
	if err != nil {
		return fmt.Errorf("login button not found")
	}

	time.Sleep(a.randDelay(800, 1500))
	loginBtn.MustClick()

	a.logger.Info("Login form submitted")
	time.Sleep(8 * time.Second)

	if a.HasSecurityCheckpoint(page) {
		return fmt.Errorf("2FA / captcha required after login")
	}

	if !a.IsLoggedIn(page) {
		return fmt.Errorf("login verification failed")
	}

	a.logger.Info("Login successful")
	return nil
}

// ==================== ELEMENT FINDERS ====================

func (a *Authenticator) findEmailInput(page *rod.Page) (*rod.Element, error) {
	selectors := []string{
		"#username",
		"input[name='session_key']",
		"input[type='text']",
	}

	for _, sel := range selectors {
		el, err := page.Timeout(6 * time.Second).Element(sel)
		if err == nil {
			a.logger.Debugf("Email input found using selector: %s", sel)
			return el, nil
		}
	}
	return nil, fmt.Errorf("email input not found")
}

func (a *Authenticator) findPasswordInput(page *rod.Page) (*rod.Element, error) {
	selectors := []string{
		"#password",
		"input[name='session_password']",
		"input[type='password']",
	}

	for _, sel := range selectors {
		el, err := page.Timeout(6 * time.Second).Element(sel)
		if err == nil {
			a.logger.Debugf("Password input found using selector: %s", sel)
			return el, nil
		}
	}
	return nil, fmt.Errorf("password input not found")
}

// ==================== HUMAN BEHAVIOR ====================

func (a *Authenticator) humanType(page *rod.Page, el *rod.Element, text string) {
	el.MustClick()
	time.Sleep(a.randDelay(300, 700))

	for _, ch := range text {
		page.Keyboard.Type(input.Key(ch))
		time.Sleep(a.randDelay(70, 160))
	}
}

func (a *Authenticator) randomScroll(page *rod.Page) {
	scrollY := rand.Intn(400) + 200
	page.Mouse.Scroll(0, float64(scrollY), 10)
	time.Sleep(a.randDelay(800, 1500))
}

func (a *Authenticator) randDelay(min, max int) time.Duration {
	return time.Duration(min+rand.Intn(max-min)) * time.Millisecond
}

// ==================== SESSION MANAGEMENT ====================

func (a *Authenticator) SaveSession(page *rod.Page) error {
	if err := os.MkdirAll(a.sessionPath, 0755); err != nil {
		return err
	}

	cookies, err := page.Cookies([]string{"https://www.linkedin.com"})
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(a.sessionPath, "cookies.json"), data, 0600)
}

func (a *Authenticator) LoadSession(page *rod.Page) error {
	path := filepath.Join(a.sessionPath, "cookies.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cookies []*proto.NetworkCookie
	if err := json.Unmarshal(data, &cookies); err != nil {
		return err
	}

	params := make([]*proto.NetworkCookieParam, 0, len(cookies))
	for _, c := range cookies {
		params = append(params, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: c.SameSite,
		})
	}

	page.SetCookies(params)

	if err := page.Navigate("https://www.linkedin.com/feed/"); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	if !a.IsLoggedIn(page) {
		return fmt.Errorf("session cookies invalid")
	}

	return nil
}

func (a *Authenticator) ClearSession() error {
	return os.Remove(filepath.Join(a.sessionPath, "cookies.json"))
}

// ==================== STATE CHECKS ====================

func (a *Authenticator) IsLoggedIn(page *rod.Page) bool {
	// Check multiple selectors that indicate successful login
	selectors := []string{
		"nav.global-nav",
		".global-nav",
		"#global-nav",
		".feed-identity-module",
		".scaffold-layout__main",
		"div[data-test-id='main-feed']",
		".feed-shared-update-v2",
		".global-nav__me",
	}

	// Also check URL patterns
	url := page.MustInfo().URL
	if strings.Contains(url, "/feed/") ||
		strings.Contains(url, "/mynetwork/") ||
		strings.Contains(url, "/messaging/") ||
		strings.Contains(url, "/notifications/") {
		a.logger.Debug("Login detected via URL pattern")
		return true
	}

	// Check if any of the selectors exist
	for _, sel := range selectors {
		el, err := page.Timeout(2 * time.Second).Element(sel)
		if err == nil && el != nil {
			a.logger.Debugf("Login detected via selector: %s", sel)
			return true
		}
	}

	return false
}

func (a *Authenticator) HasSecurityCheckpoint(page *rod.Page) bool {
	html, _ := page.HTML()
	url := page.MustInfo().URL

	signals := []string{"captcha", "checkpoint", "challenge"}

	for _, s := range signals {
		if strings.Contains(strings.ToLower(html), s) ||
			strings.Contains(strings.ToLower(url), s) {
			a.logger.Warn("Security checkpoint detected")
			return true
		}
	}
	return false
}
