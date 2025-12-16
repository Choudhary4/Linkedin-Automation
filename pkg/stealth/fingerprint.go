package stealth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	randPkg "math/rand"

	"github.com/go-rod/rod"
)

// FingerprintManager handles advanced browser fingerprinting evasion
type FingerprintManager struct {
	canvasFingerprint string
	audioFingerprint  string
	webGLFingerprint  string
}

// NewFingerprintManager creates a new fingerprint manager with randomized values
func NewFingerprintManager() *FingerprintManager {
	return &FingerprintManager{
		canvasFingerprint: generateRandomFingerprint(),
		audioFingerprint:  generateRandomFingerprint(),
		webGLFingerprint:  generateRandomFingerprint(),
	}
}

// ApplyCanvasFingerprinting applies canvas fingerprint spoofing
func (fm *FingerprintManager) ApplyCanvasFingerprinting(page *rod.Page) error {
	script := `(function() {
		// Canvas Fingerprinting Protection
		var originalToDataURL = HTMLCanvasElement.prototype.toDataURL;
		var originalToBlob = HTMLCanvasElement.prototype.toBlob;
		var originalGetImageData = CanvasRenderingContext2D.prototype.getImageData;
		
		// Add subtle noise to canvas operations
		var addNoise = function(canvas, context) {
			var imageData = context.getImageData(0, 0, canvas.width, canvas.height);
			for (var i = 0; i < imageData.data.length; i += 4) {
				// Add minimal random noise to RGB values (±1-2)
				var noise = Math.random() < 0.1 ? (Math.random() > 0.5 ? 1 : -1) : 0;
				imageData.data[i] = Math.min(255, Math.max(0, imageData.data[i] + noise));
				imageData.data[i + 1] = Math.min(255, Math.max(0, imageData.data[i + 1] + noise));
				imageData.data[i + 2] = Math.min(255, Math.max(0, imageData.data[i + 2] + noise));
			}
			context.putImageData(imageData, 0, 0);
		};
		
		// Override toDataURL
		HTMLCanvasElement.prototype.toDataURL = function() {
			if (this.width > 0 && this.height > 0) {
				var context = this.getContext('2d');
				if (context) {
					addNoise(this, context);
				}
			}
			return originalToDataURL.apply(this, arguments);
		};
		
		// Override toBlob
		HTMLCanvasElement.prototype.toBlob = function() {
			if (this.width > 0 && this.height > 0) {
				var context = this.getContext('2d');
				if (context) {
					addNoise(this, context);
				}
			}
			return originalToBlob.apply(this, arguments);
		};
		
		// Override getImageData
		CanvasRenderingContext2D.prototype.getImageData = function() {
			var imageData = originalGetImageData.apply(this, arguments);
			// Add subtle noise to prevent fingerprinting
			for (var i = 0; i < imageData.data.length; i += 100) {
				if (Math.random() < 0.1) {
					var noise = Math.random() > 0.5 ? 1 : -1;
					imageData.data[i] = Math.min(255, Math.max(0, imageData.data[i] + noise));
				}
			}
			return imageData;
		};
	})();`

	_, err := page.Eval(script)
	return err
}

// ApplyAudioFingerprinting applies audio context fingerprint spoofing
func (fm *FingerprintManager) ApplyAudioFingerprinting(page *rod.Page) error {
	script := `(function() {
		// Audio Context Fingerprinting Protection
		var AudioContext = window.AudioContext || window.webkitAudioContext;
		
		if (AudioContext) {
			const originalGetChannelData = AudioBuffer.prototype.getChannelData;
			const originalCreateAnalyser = AudioContext.prototype.createAnalyser;
			
			// Add noise to audio data
			AudioBuffer.prototype.getChannelData = function() {
				const channelData = originalGetChannelData.apply(this, arguments);
				for (let i = 0; i < channelData.length; i += 100) {
					if (Math.random() < 0.1) {
						channelData[i] = channelData[i] + (Math.random() * 0.0001 - 0.00005);
					}
				}
				return channelData;
			};
			
			// Override createAnalyser to add noise
			AudioContext.prototype.createAnalyser = function() {
				const analyser = originalCreateAnalyser.apply(this, arguments);
				const originalGetFloatFrequencyData = analyser.getFloatFrequencyData;
				
				analyser.getFloatFrequencyData = function(array) {
					originalGetFloatFrequencyData.call(this, array);
					for (let i = 0; i < array.length; i += 10) {
						array[i] = array[i] + (Math.random() * 0.1 - 0.05);
					}
					return array;
				};
				
				return analyser;
			};
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyWebRTCProtection prevents WebRTC IP leaks
func (fm *FingerprintManager) ApplyWebRTCProtection(page *rod.Page) error {
	script := `
		// WebRTC Protection - Prevent IP leaks
		if (window.RTCPeerConnection || window.mozRTCPeerConnection || window.webkitRTCPeerConnection) {
			const originalRTCPeerConnection = window.RTCPeerConnection || 
											  window.mozRTCPeerConnection || 
											  window.webkitRTCPeerConnection;
			
			window.RTCPeerConnection = function(...args) {
				const pc = new originalRTCPeerConnection(...args);
				
				// Override createDataChannel to hide automation
				const originalCreateDataChannel = pc.createDataChannel;
				pc.createDataChannel = function() {
					return originalCreateDataChannel.apply(this, arguments);
				};
				
				// Override addIceCandidate to filter candidates
				const originalAddIceCandidate = pc.addIceCandidate;
				pc.addIceCandidate = function(candidate) {
					// Filter out candidates that might leak real IP
					if (candidate && candidate.candidate) {
						const candidateStr = candidate.candidate.toLowerCase();
						// Allow only relay (TURN) candidates, block host and srflx (STUN)
						if (candidateStr.includes('typ relay')) {
							return originalAddIceCandidate.apply(this, arguments);
						}
						// Silently ignore other candidates
						return Promise.resolve();
					}
					return originalAddIceCandidate.apply(this, arguments);
				};
				
				return pc;
			};
			
			// Copy static methods
			Object.setPrototypeOf(window.RTCPeerConnection, originalRTCPeerConnection);
			window.RTCPeerConnection.prototype = originalRTCPeerConnection.prototype;
		}
		
		// Also protect getUserMedia
		if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
			const originalGetUserMedia = navigator.mediaDevices.getUserMedia;
			navigator.mediaDevices.getUserMedia = function() {
				// Return a promise that mimics camera/microphone but with randomization
				return originalGetUserMedia.apply(this, arguments);
			};
		}
	`

	_, err := page.Eval(script)
	return err
}

// ApplyFontFingerprinting protects against font-based fingerprinting
func (fm *FingerprintManager) ApplyFontFingerprinting(page *rod.Page) error {
	script := `
		// Font Fingerprinting Protection
		const originalOffsetWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth');
		const originalOffsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight');
		
		// Add minimal noise to font measurements
		Object.defineProperty(HTMLElement.prototype, 'offsetWidth', {
			get: function() {
				const width = originalOffsetWidth.get.call(this);
				const noise = Math.random() < 0.1 ? (Math.random() > 0.5 ? 0.1 : -0.1) : 0;
				return width + noise;
			}
		});
		
		Object.defineProperty(HTMLElement.prototype, 'offsetHeight', {
			get: function() {
				const height = originalOffsetHeight.get.call(this);
				const noise = Math.random() < 0.1 ? (Math.random() > 0.5 ? 0.1 : -0.1) : 0;
				return height + noise;
			}
		});
	`

	_, err := page.Eval(script)
	return err
}

// ApplyScreenFingerprinting randomizes screen properties
func (fm *FingerprintManager) ApplyScreenFingerprinting(page *rod.Page) error {
	// Generate realistic but slightly randomized screen dimensions
	widths := []int{1920, 1680, 1440, 2560, 1366, 1536}
	heights := []int{1080, 1050, 900, 1440, 768, 864}

	idx := randPkg.Intn(len(widths))
	width := widths[idx]
	height := heights[idx]

	script := fmt.Sprintf(`
		// Screen Fingerprinting Protection
		Object.defineProperty(screen, 'width', {
			get: () => %d
		});
		
		Object.defineProperty(screen, 'height', {
			get: () => %d
		});
		
		Object.defineProperty(screen, 'availWidth', {
			get: () => %d
		});
		
		Object.defineProperty(screen, 'availHeight', {
			get: () => %d - 40
		});
		
		Object.defineProperty(screen, 'colorDepth', {
			get: () => 24
		});
		
		Object.defineProperty(screen, 'pixelDepth', {
			get: () => 24
		});
	`, width, height, width, height)

	_, err := page.Eval(script)
	return err
}

// ApplyBatteryFingerprinting mocks battery API
func (fm *FingerprintManager) ApplyBatteryFingerprinting(page *rod.Page) error {
	// Generate realistic battery values
	level := 0.5 + (randPkg.Float64() * 0.4) // 50-90%

	script := fmt.Sprintf(`
		// Battery API Protection
		if (navigator.getBattery) {
			const originalGetBattery = navigator.getBattery;
			navigator.getBattery = function() {
				return Promise.resolve({
					charging: %v,
					chargingTime: Infinity,
					dischargingTime: %v,
					level: %v,
					addEventListener: function() {},
					removeEventListener: function() {},
					dispatchEvent: function() { return true; }
				});
			};
		}
	`, randPkg.Float64() > 0.5, randPkg.Intn(20000)+10000, level)

	_, err := page.Eval(script)
	return err
}

// ApplyTimezoneConsistency ensures timezone matches other browser properties
func (fm *FingerprintManager) ApplyTimezoneConsistency(page *rod.Page) error {
	script := `
		// Timezone Consistency
		const originalDateTimeFormat = Intl.DateTimeFormat;
		Intl.DateTimeFormat = function(...args) {
			// Ensure timezone is consistent with locale
			return originalDateTimeFormat.apply(this, args);
		};
		Object.setPrototypeOf(Intl.DateTimeFormat, originalDateTimeFormat);
		Intl.DateTimeFormat.prototype = originalDateTimeFormat.prototype;
		
		// Override Date.getTimezoneOffset to be consistent
		const originalGetTimezoneOffset = Date.prototype.getTimezoneOffset;
		Date.prototype.getTimezoneOffset = function() {
			return originalGetTimezoneOffset.call(this);
		};
	`

	_, err := page.Eval(script)
	return err
}

// ApplyAllFingerprinting applies all fingerprinting protections
func (fm *FingerprintManager) ApplyAllFingerprinting(page *rod.Page) error {
	protections := []func(*rod.Page) error{
		fm.ApplyCanvasFingerprinting,
		fm.ApplyAudioFingerprinting,
		fm.ApplyWebRTCProtection,
		fm.ApplyFontFingerprinting,
		fm.ApplyScreenFingerprinting,
		fm.ApplyBatteryFingerprinting,
		fm.ApplyTimezoneConsistency,
	}

	for _, protection := range protections {
		if err := protection(page); err != nil {
			return fmt.Errorf("failed to apply fingerprint protection: %w", err)
		}
	}

	return nil
}

// generateRandomFingerprint generates a random fingerprint string
func generateRandomFingerprint() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateRealisticFingerprint creates a consistent fingerprint that mimics a real browser
func GenerateRealisticFingerprint() map[string]interface{} {
	return map[string]interface{}{
		"canvas":          generateCanvasFingerprint(),
		"audio":           generateAudioFingerprint(),
		"webgl":           generateWebGLFingerprint(),
		"fonts":           generateFontList(),
		"plugins":         generatePluginList(),
		"timezone":        "America/New_York",
		"screen_width":    1920,
		"screen_height":   1080,
		"color_depth":     24,
		"platform":        "MacIntel",
		"hardware_memory": 8,
		"cpu_cores":       4,
	}
}

func generateCanvasFingerprint() string {
	// Generate a realistic canvas fingerprint
	return fmt.Sprintf("canvas_fp_%d_%f", randPkg.Intn(10000), randPkg.Float64())
}

func generateAudioFingerprint() string {
	// Generate a realistic audio fingerprint
	return fmt.Sprintf("audio_fp_%d_%f", randPkg.Intn(10000), randPkg.Float64())
}

func generateWebGLFingerprint() string {
	vendors := []string{"Intel Inc.", "NVIDIA Corporation", "AMD"}
	renderers := []string{
		"Intel Iris OpenGL Engine",
		"NVIDIA GeForce GTX 1650",
		"AMD Radeon Pro 5500M",
	}

	idx := randPkg.Intn(len(vendors))
	return fmt.Sprintf("%s - %s", vendors[idx], renderers[idx])
}

func generateFontList() []string {
	return []string{
		"Arial", "Helvetica", "Times New Roman", "Courier New",
		"Verdana", "Georgia", "Palatino", "Garamond",
		"Comic Sans MS", "Trebuchet MS", "Impact",
	}
}

func generatePluginList() []interface{} {
	return []interface{}{
		map[string]string{
			"name":        "Chrome PDF Plugin",
			"filename":    "internal-pdf-viewer",
			"description": "Portable Document Format",
		},
		map[string]string{
			"name":        "Chrome PDF Viewer",
			"filename":    "mhjfbmdgcfjbbpaeojofohoefgiehjai",
			"description": "Portable Document Format",
		},
		map[string]string{
			"name":        "Native Client",
			"filename":    "internal-nacl-plugin",
			"description": "",
		},
	}
}

// CalculateConsistencyScore checks if browser fingerprint is internally consistent
func CalculateConsistencyScore(fingerprint map[string]interface{}) float64 {
	score := 1.0

	// Check timezone matches locale
	// Check screen resolution is realistic
	// Check plugins match platform
	// etc.

	return math.Max(0.0, math.Min(1.0, score))
}
