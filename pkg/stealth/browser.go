package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// SetupStealthBrowser configures a browser with anti-detection measures
func SetupStealthBrowser(headless bool, slowMotion int) (*rod.Browser, error) {
	// Setup launcher with extensive anti-detection flags
	l := launcher.New().
		Headless(headless).
		// Core anti-detection
		Set("disable-blink-features", "AutomationControlled").
		Set("exclude-switches", "enable-automation").
		Set("disable-infobars").
		// Additional stealth
		Set("disable-web-security").
		Set("disable-features", "IsolateOrigins,site-per-process").
		Set("allow-running-insecure-content").
		Set("disable-ipc-flooding-protection").
		Set("disable-dev-shm-usage").
		// Window and performance
		Set("window-size", "1920,1080").
		Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Launch browser
	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	// Connect to browser with stealth settings
	browser := rod.New().
		ControlURL(url).
		SlowMotion(time.Duration(slowMotion) * time.Millisecond)

	err = browser.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}

	return browser, nil
}

// ApplyStealthSettings applies stealth JavaScript to mask automation
func ApplyStealthSettings(page *rod.Page) error {
	// Apply stealth plugin to evade detection
	_ = page // Stealth is applied via JavaScript below

	// Additional fingerprint masking
	_, err := page.Eval(`() => {
		// Override navigator.webdriver
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});

		// Override navigator properties for more realistic fingerprint
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{
					0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "internal-pdf-viewer",
					length: 1,
					name: "Chrome PDF Plugin"
				},
				{
					0: {type: "application/pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
					length: 1,
					name: "Chrome PDF Viewer"
				},
				{
					0: {type: "application/x-nacl", suffixes: "", description: "Native Client Executable"},
					1: {type: "application/x-pnacl", suffixes: "", description: "Portable Native Client Executable"},
					description: "",
					filename: "internal-nacl-plugin",
					length: 2,
					name: "Native Client"
				}
			]
		});

		// Override navigator.languages
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en']
		});

		// Override screen properties
		Object.defineProperty(screen, 'colorDepth', {
			get: () => 24
		});

		// Override permissions
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({state: Notification.permission}) :
				originalQuery(parameters)
		);

		// Remove automation-related properties
		delete navigator.__proto__.webdriver;

		// Override chrome runtime
		window.chrome = {
			runtime: {},
			loadTimes: function() {},
			csi: function() {},
			app: {}
		};

		// Mock realistic user agent data
		if (navigator.userAgentData) {
			Object.defineProperty(navigator, 'userAgentData', {
				get: () => ({
					brands: [
						{brand: "Not_A Brand", version: "8"},
						{brand: "Chromium", version: "120"},
						{brand: "Google Chrome", version: "120"}
					],
					mobile: false,
					platform: "macOS"
				})
			});
		}
		
		// Override WebGL vendor to avoid headless detection
		const getParameter = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(parameter) {
			if (parameter === 37445) {
				return 'Intel Inc.';
			}
			if (parameter === 37446) {
				return 'Intel Iris OpenGL Engine';
			}
			return getParameter.call(this, parameter);
		};
		
		// Hide automation in toString
		['language', 'languages', 'webdriver', 'plugins'].forEach(prop => {
			const originalDescriptor = Object.getOwnPropertyDescriptor(Navigator.prototype, prop);
			if (originalDescriptor && originalDescriptor.get) {
				const originalGet = originalDescriptor.get;
				Object.defineProperty(Navigator.prototype, prop, {
					get: new Proxy(originalGet, {
						apply(target, thisArg, args) {
							const ret = Reflect.apply(target, thisArg, args);
							return ret;
						}
					})
				});
			}
		});
	}`)
	return err
}

// RandomViewport sets a random but realistic viewport size
func RandomViewport(page *rod.Page) error {
	viewports := []struct {
		width  int
		height int
	}{
		{1920, 1080}, // Full HD
		{1680, 1050}, // WSXGA+
		{1600, 900},  // HD+
		{1440, 900},  // WXGA+
		{1366, 768},  // Common laptop
		{1536, 864},  // Common laptop
		{1280, 720},  // HD
	}

	rand.Seed(time.Now().UnixNano())
	viewport := viewports[rand.Intn(len(viewports))]

	return page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width:  viewport.width,
		Height: viewport.height,
	})
}

// RandomUserAgent returns a random realistic user agent string
func RandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	}

	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// SetRandomUserAgent sets a random user agent for the page
func SetRandomUserAgent(page *rod.Page) error {
	userAgent := RandomUserAgent()
	return page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: userAgent,
	})
}

// DisableAutomationFlags disables various automation detection flags
func DisableAutomationFlags(page *rod.Page) error {
	// This is handled by the stealth plugin and additional eval scripts
	// But we can add more specific overrides here
	_, err := page.Eval(`() => {
		// Additional automation flag removal
		const originalNavigator = window.navigator;
		
		// Create a proxy to intercept navigator property access
		const navigatorProxy = new Proxy(originalNavigator, {
			get: function(target, prop) {
				if (prop === 'webdriver') {
					return undefined;
				}
				return target[prop];
			}
		});
		
		// Try to replace navigator (may not work in all contexts)
		try {
			Object.defineProperty(window, 'navigator', {
				value: navigatorProxy,
				configurable: false,
				writable: false
			});
		} catch (e) {
			// Navigator is already defined
		}
	}`)
	return err
}
