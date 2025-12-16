package stealth

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// UltraStealthManager provides maximum anti-detection capabilities
type UltraStealthManager struct {
	hardwareProfile HardwareProfile
}

// HardwareProfile represents realistic hardware configuration
type HardwareProfile struct {
	DeviceMemory    int
	HardwareCores   int
	MaxTouchPoints  int
	Platform        string
	Vendor          string
	Renderer        string
	Architecture    string
	Model           string
	PlatformVersion string
}

// NewUltraStealthManager creates a new ultra stealth manager
func NewUltraStealthManager() *UltraStealthManager {
	return &UltraStealthManager{
		hardwareProfile: generateRealisticHardwareProfile(),
	}
}

// generateRealisticHardwareProfile creates a realistic hardware profile
func generateRealisticHardwareProfile() HardwareProfile {
	profiles := []HardwareProfile{
		{
			DeviceMemory:    8,
			HardwareCores:   4,
			MaxTouchPoints:  0,
			Platform:        "MacIntel",
			Vendor:          "Apple Inc.",
			Renderer:        "Apple M1",
			Architecture:    "arm64",
			Model:           "",
			PlatformVersion: "14.0.0",
		},
		{
			DeviceMemory:    16,
			HardwareCores:   8,
			MaxTouchPoints:  0,
			Platform:        "MacIntel",
			Vendor:          "Intel Inc.",
			Renderer:        "Intel Iris Plus Graphics 655",
			Architecture:    "x86_64",
			Model:           "",
			PlatformVersion: "13.5.0",
		},
		{
			DeviceMemory:    8,
			HardwareCores:   4,
			MaxTouchPoints:  0,
			Platform:        "Win32",
			Vendor:          "Google Inc. (NVIDIA)",
			Renderer:        "ANGLE (NVIDIA, NVIDIA GeForce GTX 1650 Direct3D11 vs_5_0 ps_5_0, D3D11)",
			Architecture:    "x86_64",
			Model:           "",
			PlatformVersion: "10.0.0",
		},
		{
			DeviceMemory:    16,
			HardwareCores:   6,
			MaxTouchPoints:  0,
			Platform:        "Win32",
			Vendor:          "Google Inc. (AMD)",
			Renderer:        "ANGLE (AMD, AMD Radeon RX 580 Series Direct3D11 vs_5_0 ps_5_0, D3D11)",
			Architecture:    "x86_64",
			Model:           "",
			PlatformVersion: "10.0.0",
		},
	}

	return profiles[rand.Intn(len(profiles))]
}

// ApplyUltraStealth applies all ultra stealth techniques
func (usm *UltraStealthManager) ApplyUltraStealth(page *rod.Page) error {
	// Apply in specific order for maximum effectiveness
	techniques := []struct {
		name string
		fn   func(*rod.Page) error
	}{
		{"CDP Evasion", usm.ApplyCDPEvasion},
		{"Hardware Fingerprint", usm.ApplyHardwareFingerprint},
		{"Navigator Override", usm.ApplyNavigatorOverride},
		{"Window Properties", usm.ApplyWindowProperties},
		{"Console Protection", usm.ApplyConsoleProtection},
		{"Iframe Protection", usm.ApplyIframeProtection},
		{"Performance API", usm.ApplyPerformanceProtection},
		{"Speech Recognition", usm.ApplySpeechProtection},
		{"Clipboard API", usm.ApplyClipboardProtection},
		{"Network Information", usm.ApplyNetworkInfoProtection},
		{"MediaDevices", usm.ApplyMediaDevicesProtection},
		{"Document Properties", usm.ApplyDocumentProtection},
	}

	for _, technique := range techniques {
		if err := technique.fn(page); err != nil {
			return fmt.Errorf("failed to apply %s: %w", technique.name, err)
		}
	}

	return nil
}

// ApplyCDPEvasion prevents detection of Chrome DevTools Protocol usage
func (usm *UltraStealthManager) ApplyCDPEvasion(page *rod.Page) error {
	script := `
		// Prevent CDP detection
		const cdpProps = [
			'__webdriver_evaluate',
			'__selenium_evaluate', 
			'__webdriver_script_function',
			'__webdriver_script_func',
			'__webdriver_script_fn',
			'__fxdriver_evaluate',
			'__driver_unwrapped',
			'__webdriver_unwrapped',
			'__driver_evaluate',
			'__selenium_unwrapped',
			'__fxdriver_unwrapped',
			'_Selenium_IDE_Recorder',
			'_selenium',
			'calledSelenium',
			'$cdc_asdjflasutopfhvcZLmcfl_',
			'$chrome_asyncScriptInfo',
			'__$webdriverAsyncExecutor',
			'webdriver',
			'__webdriver_script_fn',
			'__lastWatirAlert',
			'__lastWatirConfirm',
			'__lastWatirPrompt',
			'$chrome_asyncScriptInfo'
		];
		
		// Delete from window
		cdpProps.forEach(prop => {
			try {
				if (prop in window) {
					delete window[prop];
				}
			} catch(e) {}
		});
		
		// Delete from document
		cdpProps.forEach(prop => {
			try {
				if (prop in document) {
					delete document[prop];
				}
			} catch(e) {}
		});
		
		// Override cdc_ detection (Chrome DevTools Protocol marker)
		Object.defineProperty(window, 'cdc_adoQpoasnfa76pfcZLmcfl_Array', {
			get: () => undefined,
			configurable: false
		});
		
		Object.defineProperty(window, 'cdc_adoQpoasnfa76pfcZLmcfl_Promise', {
			get: () => undefined,
			configurable: false
		});
		
		Object.defineProperty(window, 'cdc_adoQpoasnfa76pfcZLmcfl_Symbol', {
			get: () => undefined,
			configurable: false
		});
		
		// Prevent runtime.enable detection
		if (window.chrome && window.chrome.runtime) {
			const originalRuntime = window.chrome.runtime;
			window.chrome.runtime = new Proxy(originalRuntime, {
				get: function(target, name) {
					if (name === 'id') return undefined;
					return target[name];
				}
			});
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyHardwareFingerprint spoofs hardware-related browser properties
func (usm *UltraStealthManager) ApplyHardwareFingerprint(page *rod.Page) error {
	script := fmt.Sprintf(`
		// Hardware Concurrency (CPU cores)
		Object.defineProperty(navigator, 'hardwareConcurrency', {
			get: () => %d,
			configurable: true
		});
		
		// Device Memory
		Object.defineProperty(navigator, 'deviceMemory', {
			get: () => %d,
			configurable: true
		});
		
		// Max Touch Points
		Object.defineProperty(navigator, 'maxTouchPoints', {
			get: () => %d,
			configurable: true
		});
		
		// Platform
		Object.defineProperty(navigator, 'platform', {
			get: () => '%s',
			configurable: true
		});
		
		// Vendor
		Object.defineProperty(navigator, 'vendor', {
			get: () => 'Google Inc.',
			configurable: true
		});
		
		// Product
		Object.defineProperty(navigator, 'product', {
			get: () => 'Gecko',
			configurable: true
		});
		
		// App Version
		Object.defineProperty(navigator, 'appVersion', {
			get: () => navigator.userAgent.replace('Mozilla/', ''),
			configurable: true
		});
		
		// WebGL Vendor/Renderer (advanced)
		const getParameterProxyHandler = {
			apply: function(target, thisArg, argumentsList) {
				const param = argumentsList[0];
				// UNMASKED_VENDOR_WEBGL
				if (param === 37445) {
					return '%s';
				}
				// UNMASKED_RENDERER_WEBGL
				if (param === 37446) {
					return '%s';
				}
				return Reflect.apply(target, thisArg, argumentsList);
			}
		};
		
		// Apply to both WebGL contexts
		['WebGLRenderingContext', 'WebGL2RenderingContext'].forEach(ctx => {
			if (window[ctx]) {
				const originalGetParameter = window[ctx].prototype.getParameter;
				window[ctx].prototype.getParameter = new Proxy(originalGetParameter, getParameterProxyHandler);
			}
		});
	`, usm.hardwareProfile.HardwareCores, usm.hardwareProfile.DeviceMemory,
		usm.hardwareProfile.MaxTouchPoints, usm.hardwareProfile.Platform,
		usm.hardwareProfile.Vendor, usm.hardwareProfile.Renderer)

	_, err := page.Eval(script)
	return err
}

// ApplyNavigatorOverride applies comprehensive navigator overrides
func (usm *UltraStealthManager) ApplyNavigatorOverride(page *rod.Page) error {
	script := `
		// Connection type
		if (navigator.connection) {
			Object.defineProperty(navigator.connection, 'rtt', {
				get: () => 50 + Math.floor(Math.random() * 50),
				configurable: true
			});
			
			Object.defineProperty(navigator.connection, 'downlink', {
				get: () => 5 + Math.random() * 5,
				configurable: true
			});
			
			Object.defineProperty(navigator.connection, 'effectiveType', {
				get: () => '4g',
				configurable: true
			});
			
			Object.defineProperty(navigator.connection, 'saveData', {
				get: () => false,
				configurable: true
			});
		}
		
		// DoNotTrack
		Object.defineProperty(navigator, 'doNotTrack', {
			get: () => null,
			configurable: true
		});
		
		// Cookie Enabled
		Object.defineProperty(navigator, 'cookieEnabled', {
			get: () => true,
			configurable: true
		});
		
		// Java Enabled
		Object.defineProperty(navigator, 'javaEnabled', {
			value: () => false,
			configurable: true
		});
		
		// PDF Viewer Enabled
		Object.defineProperty(navigator, 'pdfViewerEnabled', {
			get: () => true,
			configurable: true
		});
		
		// Bluetooth
		if (navigator.bluetooth) {
			Object.defineProperty(navigator, 'bluetooth', {
				get: () => undefined,
				configurable: true
			});
		}
		
		// USB
		if (navigator.usb) {
			Object.defineProperty(navigator, 'usb', {
				get: () => undefined,
				configurable: true
			});
		}
		
		// Serial
		if (navigator.serial) {
			Object.defineProperty(navigator, 'serial', {
				get: () => undefined,
				configurable: true
			});
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyWindowProperties protects window-level properties
func (usm *UltraStealthManager) ApplyWindowProperties(page *rod.Page) error {
	script := `
		// Outer dimensions (should match inner when maximized)
		Object.defineProperty(window, 'outerWidth', {
			get: () => window.innerWidth,
			configurable: true
		});
		
		Object.defineProperty(window, 'outerHeight', {
			get: () => window.innerHeight + 85, // Account for browser chrome
			configurable: true
		});
		
		// DevTools detection prevention
		let devtoolsOpen = false;
		const threshold = 160;
		
		const checkDevTools = () => {
			// Always return false to prevent detection
			return false;
		};
		
		// Override console methods that can detect DevTools
		const originalConsoleLog = console.log;
		console.log = function() {
			// Filter out devtools detection attempts
			return originalConsoleLog.apply(console, arguments);
		};
		
		// Prevent Firebug detection
		Object.defineProperty(window, 'Firebug', {
			get: () => undefined,
			configurable: true
		});
		
		// History length randomization
		const originalHistoryLength = Object.getOwnPropertyDescriptor(History.prototype, 'length');
		Object.defineProperty(History.prototype, 'length', {
			get: function() {
				const realLength = originalHistoryLength.get.call(this);
				return Math.max(realLength, 2 + Math.floor(Math.random() * 5));
			},
			configurable: true
		});
		
		// OffscreenCanvas
		if (window.OffscreenCanvas) {
			const OriginalOffscreenCanvas = window.OffscreenCanvas;
			window.OffscreenCanvas = function(width, height) {
				const canvas = new OriginalOffscreenCanvas(width, height);
				return canvas;
			};
			window.OffscreenCanvas.prototype = OriginalOffscreenCanvas.prototype;
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyConsoleProtection prevents detection via console methods
func (usm *UltraStealthManager) ApplyConsoleProtection(page *rod.Page) error {
	script := `
		// Protect console methods from detection
		const consoleMethods = ['log', 'warn', 'error', 'info', 'debug', 'table', 'trace'];
		
		consoleMethods.forEach(method => {
			const original = console[method];
			console[method] = function() {
				// Check if this is a detection attempt
				const args = Array.from(arguments);
				const strArgs = args.map(a => String(a)).join('');
				
				// Block known detection patterns
				if (strArgs.includes('%c') && strArgs.includes('font-size')) {
					return; // Block image-based console detection
				}
				
				return original.apply(console, arguments);
			};
			
			// Preserve toString
			console[method].toString = () => 'function ' + method + '() { [native code] }';
		});
		
		// Protect console.profile
		console.profile = function() {};
		console.profileEnd = function() {};
		console.profile.toString = () => 'function profile() { [native code] }';
		console.profileEnd.toString = () => 'function profileEnd() { [native code] }';
	`

	_, err := page.Eval(script)
	return err
}

// ApplyIframeProtection prevents iframe-based fingerprinting
func (usm *UltraStealthManager) ApplyIframeProtection(page *rod.Page) error {
	script := `
		// Protect contentWindow from cross-origin detection
		const originalContentWindow = Object.getOwnPropertyDescriptor(HTMLIFrameElement.prototype, 'contentWindow');
		
		Object.defineProperty(HTMLIFrameElement.prototype, 'contentWindow', {
			get: function() {
				const win = originalContentWindow.get.call(this);
				if (win) {
					// Apply same protections to iframe
					try {
						Object.defineProperty(win.navigator, 'webdriver', {
							get: () => undefined,
							configurable: true
						});
					} catch(e) {}
				}
				return win;
			},
			configurable: true
		});
		
		// Srcdoc iframe protection
		const originalSrcDoc = Object.getOwnPropertyDescriptor(HTMLIFrameElement.prototype, 'srcdoc');
		if (originalSrcDoc) {
			Object.defineProperty(HTMLIFrameElement.prototype, 'srcdoc', {
				get: originalSrcDoc.get,
				set: function(value) {
					// Inject stealth script into srcdoc
					const stealthScript = '<script>Object.defineProperty(navigator,"webdriver",{get:()=>undefined});</script>';
					originalSrcDoc.set.call(this, stealthScript + value);
				},
				configurable: true
			});
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyPerformanceProtection protects Performance API from fingerprinting
func (usm *UltraStealthManager) ApplyPerformanceProtection(page *rod.Page) error {
	script := `
		// Add noise to performance.now()
		const originalNow = performance.now.bind(performance);
		performance.now = function() {
			const now = originalNow();
			// Add micro-noise to prevent timing attacks
			return now + (Math.random() * 0.1);
		};
		performance.now.toString = () => 'function now() { [native code] }';
		
		// Protect memory API
		if (performance.memory) {
			const randomOffset = Math.floor(Math.random() * 1000000);
			Object.defineProperty(performance, 'memory', {
				get: () => ({
					jsHeapSizeLimit: 2172649472 + randomOffset,
					totalJSHeapSize: 19573456 + randomOffset,
					usedJSHeapSize: 16319312 + randomOffset
				}),
				configurable: true
			});
		}
		
		// Protect navigation timing
		const originalGetEntriesByType = performance.getEntriesByType.bind(performance);
		performance.getEntriesByType = function(type) {
			const entries = originalGetEntriesByType(type);
			// Add slight randomization to timing entries
			return entries.map(entry => {
				if (entry.entryType === 'navigation' || entry.entryType === 'resource') {
					// Clone and add noise
					const cloned = {};
					for (const key in entry) {
						if (typeof entry[key] === 'number') {
							cloned[key] = entry[key] + (Math.random() * 0.5);
						} else {
							cloned[key] = entry[key];
						}
					}
					return cloned;
				}
				return entry;
			});
		};
	`

	_, err := page.Eval(script)
	return err
}

// ApplySpeechProtection protects Speech Recognition API
func (usm *UltraStealthManager) ApplySpeechProtection(page *rod.Page) error {
	script := `
		// Protect Speech Recognition
		if (window.SpeechRecognition || window.webkitSpeechRecognition) {
			const OriginalSpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
			
			window.SpeechRecognition = function() {
				return new OriginalSpeechRecognition();
			};
			window.SpeechRecognition.prototype = OriginalSpeechRecognition.prototype;
			
			if (window.webkitSpeechRecognition) {
				window.webkitSpeechRecognition = window.SpeechRecognition;
			}
		}
		
		// Speech Synthesis voices randomization
		if (window.speechSynthesis) {
			const originalGetVoices = speechSynthesis.getVoices.bind(speechSynthesis);
			speechSynthesis.getVoices = function() {
				const voices = originalGetVoices();
				// Return voices but with slight shuffle to prevent fingerprinting
				return voices.sort(() => 0.5 - Math.random()).slice(0, voices.length);
			};
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyClipboardProtection protects Clipboard API
func (usm *UltraStealthManager) ApplyClipboardProtection(page *rod.Page) error {
	script := `
		// Protect clipboard from fingerprinting reads
		if (navigator.clipboard) {
			const originalReadText = navigator.clipboard.readText;
			navigator.clipboard.readText = async function() {
				// Only allow if user gesture
				try {
					return await originalReadText.call(this);
				} catch(e) {
					return '';
				}
			};
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyNetworkInfoProtection protects Network Information API
func (usm *UltraStealthManager) ApplyNetworkInfoProtection(page *rod.Page) error {
	script := `
		// Protect Network Information API
		if (navigator.connection) {
			const connectionProps = {
				downlink: 10,
				effectiveType: '4g',
				rtt: 50,
				saveData: false,
				type: 'wifi'
			};
			
			Object.keys(connectionProps).forEach(prop => {
				try {
					Object.defineProperty(navigator.connection, prop, {
						get: () => connectionProps[prop],
						configurable: true
					});
				} catch(e) {}
			});
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyMediaDevicesProtection protects MediaDevices enumeration
func (usm *UltraStealthManager) ApplyMediaDevicesProtection(page *rod.Page) error {
	script := `
		// Protect MediaDevices enumeration
		if (navigator.mediaDevices && navigator.mediaDevices.enumerateDevices) {
			const originalEnumerateDevices = navigator.mediaDevices.enumerateDevices.bind(navigator.mediaDevices);
			
			navigator.mediaDevices.enumerateDevices = async function() {
				const devices = await originalEnumerateDevices();
				
				// Return generic device labels to prevent fingerprinting
				return devices.map((device, index) => ({
					deviceId: device.deviceId,
					groupId: device.groupId,
					kind: device.kind,
					label: device.label ? (device.kind === 'audioinput' ? 'Microphone' : 
							device.kind === 'audiooutput' ? 'Speaker' : 
							device.kind === 'videoinput' ? 'Camera' : device.label) : ''
				}));
			};
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyDocumentProtection protects document-level properties
func (usm *UltraStealthManager) ApplyDocumentProtection(page *rod.Page) error {
	script := `
		// Document domain protection
		Object.defineProperty(document, 'domain', {
			get: () => location.hostname,
			set: () => {},
			configurable: true
		});
		
		// Referrer protection
		Object.defineProperty(document, 'referrer', {
			get: function() {
				// Return realistic referrer
				if (location.href.includes('linkedin.com')) {
					return 'https://www.google.com/';
				}
				return '';
			},
			configurable: true
		});
		
		// Hidden/Visibility state
		Object.defineProperty(document, 'hidden', {
			get: () => false,
			configurable: true
		});
		
		Object.defineProperty(document, 'visibilityState', {
			get: () => 'visible',
			configurable: true
		});
		
		// Prevent document.hasFocus detection
		document.hasFocus = function() {
			return true;
		};
		document.hasFocus.toString = () => 'function hasFocus() { [native code] }';
	`

	_, err := page.Eval(script)
	return err
}

// ApplyLinkedInSpecificProtections applies LinkedIn-specific anti-detection
func (usm *UltraStealthManager) ApplyLinkedInSpecificProtections(page *rod.Page) error {
	script := `
		// Block LinkedIn's fingerprinting scripts
		const originalFetch = window.fetch;
		window.fetch = async function(url, options) {
			const urlStr = typeof url === 'string' ? url : url.url;
			
			// Block known fingerprinting endpoints
			if (urlStr && (
				urlStr.includes('li/track') ||
				urlStr.includes('platform-telemetry') ||
				urlStr.includes('li/pixel') ||
				urlStr.includes('analytics')
			)) {
				return new Response('', { status: 200 });
			}
			
			return originalFetch.apply(this, arguments);
		};
		
		// Block XHR fingerprinting
		const originalXHROpen = XMLHttpRequest.prototype.open;
		XMLHttpRequest.prototype.open = function(method, url) {
			const urlStr = typeof url === 'string' ? url : url.toString();
			
			if (urlStr && (
				urlStr.includes('li/track') ||
				urlStr.includes('platform-telemetry') ||
				urlStr.includes('li/pixel')
			)) {
				// Redirect to null endpoint
				arguments[1] = 'about:blank';
			}
			
			return originalXHROpen.apply(this, arguments);
		};
		
		// Prevent beacon tracking
		const originalSendBeacon = navigator.sendBeacon;
		navigator.sendBeacon = function(url, data) {
			const urlStr = typeof url === 'string' ? url : url.toString();
			
			if (urlStr && (
				urlStr.includes('li/track') ||
				urlStr.includes('platform-telemetry')
			)) {
				return true; // Pretend it was sent
			}
			
			return originalSendBeacon.apply(this, arguments);
		};
		
		// Remove LinkedIn tracking cookies from being set
		const originalSetItem = Storage.prototype.setItem;
		Storage.prototype.setItem = function(key, value) {
			// Block known tracking storage keys
			if (key && (
				key.includes('li_') ||
				key.includes('lms_') ||
				key.includes('voyager')
			)) {
				// Allow but modify tracking data
			}
			return originalSetItem.apply(this, arguments);
		};
	`

	_, err := page.Eval(script)
	return err
}

// SetupRequestInterception intercepts and modifies network requests
func SetupRequestInterception(page *rod.Page) error {
	// Enable request interception
	err := proto.FetchEnable{
		Patterns: []*proto.FetchRequestPattern{
			{URLPattern: "*"},
		},
	}.Call(page)

	if err != nil {
		return err
	}

	go page.EachEvent(func(e *proto.FetchRequestPaused) {
		// Build new headers
		var newHeaders []*proto.FetchHeaderEntry

		// Copy existing headers
		for k, v := range e.Request.Headers {
			newHeaders = append(newHeaders, &proto.FetchHeaderEntry{
				Name:  k,
				Value: v.String(),
			})
		}

		// Add realistic headers
		additionalHeaders := map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept-Encoding":           "gzip, deflate, br",
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
			"Sec-Fetch-User":            "?1",
			"Upgrade-Insecure-Requests": "1",
		}

		for k, v := range additionalHeaders {
			newHeaders = append(newHeaders, &proto.FetchHeaderEntry{
				Name:  k,
				Value: v,
			})
		}

		// Continue request
		proto.FetchContinueRequest{
			RequestID: e.RequestID,
			Headers:   newHeaders,
		}.Call(page)
	})()

	return nil
}

// convertHeaders converts map to proto headers
func convertHeaders(headers map[string]string) []*proto.FetchHeaderEntry {
	var result []*proto.FetchHeaderEntry
	for k, v := range headers {
		result = append(result, &proto.FetchHeaderEntry{
			Name:  k,
			Value: v,
		})
	}
	return result
}

// GenerateRealisticClientHints generates realistic client hints
func GenerateRealisticClientHints() map[string]string {
	platforms := []string{"macOS", "Windows", "Linux"}
	platform := platforms[rand.Intn(len(platforms))]

	brands := []struct {
		brand   string
		version string
	}{
		{"Chromium", "120"},
		{"Google Chrome", "120"},
		{"Not_A Brand", "24"},
	}

	var brandList []string
	for _, b := range brands {
		brandList = append(brandList, fmt.Sprintf(`"%s";v="%s"`, b.brand, b.version))
	}

	return map[string]string{
		"sec-ch-ua":                   strings.Join(brandList, ", "),
		"sec-ch-ua-mobile":            "?0",
		"sec-ch-ua-platform":          fmt.Sprintf(`"%s"`, platform),
		"sec-ch-ua-platform-version":  `"14.0.0"`,
		"sec-ch-ua-arch":              `"arm"`,
		"sec-ch-ua-bitness":           `"64"`,
		"sec-ch-ua-model":             `""`,
		"sec-ch-ua-full-version-list": strings.Join(brandList, ", "),
	}
}
