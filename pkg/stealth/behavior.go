package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

// MouseBehavior simulates realistic mouse behavior
type MouseBehavior struct {
	page *rod.Page
}

// NewMouseBehavior creates a new mouse behavior simulator
func NewMouseBehavior(page *rod.Page) *MouseBehavior {
	return &MouseBehavior{page: page}
}

// BezierCurveMove moves mouse along a bezier curve (human-like movement)
func (mb *MouseBehavior) BezierCurveMove(startX, startY, endX, endY float64) error {
	// Generate control points for bezier curve
	cp1X := startX + (endX-startX)*0.25 + (rand.Float64()-0.5)*50
	cp1Y := startY + (endY-startY)*0.25 + (rand.Float64()-0.5)*50
	cp2X := startX + (endX-startX)*0.75 + (rand.Float64()-0.5)*50
	cp2Y := startY + (endY-startY)*0.75 + (rand.Float64()-0.5)*50

	// Number of steps (more steps = smoother movement)
	steps := 20 + rand.Intn(30)

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)

		// Cubic bezier formula
		x := cubicBezier(startX, cp1X, cp2X, endX, t)
		y := cubicBezier(startY, cp1Y, cp2Y, endY, t)

		mb.page.Mouse.MustMoveTo(x, y)

		// Variable delay between movements
		delay := time.Duration(5+rand.Intn(15)) * time.Millisecond
		time.Sleep(delay)
	}

	return nil
}

// cubicBezier calculates point on cubic bezier curve
func cubicBezier(p0, p1, p2, p3, t float64) float64 {
	return (1-t)*(1-t)*(1-t)*p0 +
		3*(1-t)*(1-t)*t*p1 +
		3*(1-t)*t*t*p2 +
		t*t*t*p3
}

// HumanLikeClick performs a human-like click with pre-movement
func (mb *MouseBehavior) HumanLikeClick(x, y float64) error {
	// Get current position
	currentX := 0.0 + rand.Float64()*100
	currentY := 0.0 + rand.Float64()*100

	// Move to target with bezier curve
	if err := mb.BezierCurveMove(currentX, currentY, x, y); err != nil {
		return err
	}

	// Small overshoot and correction (human behavior)
	if rand.Float64() < 0.3 {
		overshootX := x + (rand.Float64()-0.5)*10
		overshootY := y + (rand.Float64()-0.5)*10
		mb.page.Mouse.MustMoveTo(overshootX, overshootY)
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		mb.page.Mouse.MustMoveTo(x, y)
	}

	// Small delay before click
	time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)

	// Click with realistic timing
	mb.page.Mouse.MustDown("left")
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
	mb.page.Mouse.MustUp("left")

	return nil
}

// RandomMouseMovement performs random mouse movements (idle behavior)
func (mb *MouseBehavior) RandomMouseMovement(duration time.Duration) error {
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		// Random target within viewport
		targetX := 100.0 + rand.Float64()*1700
		targetY := 100.0 + rand.Float64()*900

		// Move with bezier curve
		currentX := 100.0 + rand.Float64()*1700
		currentY := 100.0 + rand.Float64()*900

		mb.BezierCurveMove(currentX, currentY, targetX, targetY)

		// Random pause
		time.Sleep(time.Duration(500+rand.Intn(2000)) * time.Millisecond)
	}

	return nil
}

// KeyboardBehavior simulates realistic keyboard behavior
type KeyboardBehavior struct {
	page *rod.Page
}

// NewKeyboardBehavior creates a new keyboard behavior simulator
func NewKeyboardBehavior(page *rod.Page) *KeyboardBehavior {
	return &KeyboardBehavior{page: page}
}

// HumanLikeType types text with human-like timing and occasional mistakes
func (kb *KeyboardBehavior) HumanLikeType(text string) error {
	for i, char := range text {
		// Occasional typo (3% chance)
		if rand.Float64() < 0.03 && i > 0 {
			// Make a typo
			wrongChar := getAdjacentKey(char)
			kb.page.InsertText(string(wrongChar))
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

			// Delete the typo
			kb.page.Keyboard.MustType(input.Backspace)
			time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)
		}

		// Type the correct character using InsertText
		kb.page.InsertText(string(char))

		// Variable delay based on character type
		var delay time.Duration
		switch {
		case char == ' ':
			delay = time.Duration(100+rand.Intn(150)) * time.Millisecond
		case char >= 'A' && char <= 'Z':
			delay = time.Duration(80+rand.Intn(120)) * time.Millisecond // Capitals take longer
		default:
			delay = time.Duration(50+rand.Intn(100)) * time.Millisecond
		}

		// Occasional longer pause (thinking)
		if rand.Float64() < 0.1 {
			delay += time.Duration(200+rand.Intn(500)) * time.Millisecond
		}

		time.Sleep(delay)
	}

	return nil
}

// getAdjacentKey returns an adjacent key for typo simulation
func getAdjacentKey(char rune) rune {
	adjacentKeys := map[rune][]rune{
		'a': {'s', 'q', 'w', 'z'},
		'b': {'v', 'g', 'h', 'n'},
		'c': {'x', 'd', 'f', 'v'},
		'd': {'s', 'e', 'r', 'f', 'c', 'x'},
		'e': {'w', 's', 'd', 'r'},
		'f': {'d', 'r', 't', 'g', 'v', 'c'},
		'g': {'f', 't', 'y', 'h', 'b', 'v'},
		'h': {'g', 'y', 'u', 'j', 'n', 'b'},
		'i': {'u', 'j', 'k', 'o'},
		'j': {'h', 'u', 'i', 'k', 'm', 'n'},
		'k': {'j', 'i', 'o', 'l', 'm'},
		'l': {'k', 'o', 'p'},
		'm': {'n', 'j', 'k'},
		'n': {'b', 'h', 'j', 'm'},
		'o': {'i', 'k', 'l', 'p'},
		'p': {'o', 'l'},
		'q': {'w', 'a'},
		'r': {'e', 'd', 'f', 't'},
		's': {'a', 'w', 'e', 'd', 'x', 'z'},
		't': {'r', 'f', 'g', 'y'},
		'u': {'y', 'h', 'j', 'i'},
		'v': {'c', 'f', 'g', 'b'},
		'w': {'q', 'a', 's', 'e'},
		'x': {'z', 's', 'd', 'c'},
		'y': {'t', 'g', 'h', 'u'},
		'z': {'a', 's', 'x'},
	}

	lowerChar := char
	if char >= 'A' && char <= 'Z' {
		lowerChar = char + 32
	}

	if adjacent, ok := adjacentKeys[lowerChar]; ok && len(adjacent) > 0 {
		return adjacent[rand.Intn(len(adjacent))]
	}

	return char
}

// ScrollBehavior simulates realistic scrolling
type ScrollBehavior struct {
	page *rod.Page
}

// NewScrollBehavior creates a new scroll behavior simulator
func NewScrollBehavior(page *rod.Page) *ScrollBehavior {
	return &ScrollBehavior{page: page}
}

// HumanLikeScroll performs human-like scrolling with variable speeds
func (sb *ScrollBehavior) HumanLikeScroll(totalDistance float64, direction int) error {
	// direction: 1 = down, -1 = up
	scrolled := 0.0

	for scrolled < totalDistance {
		// Variable scroll amount (humans don't scroll consistently)
		scrollAmount := 50.0 + rand.Float64()*150

		if scrolled+scrollAmount > totalDistance {
			scrollAmount = totalDistance - scrolled
		}

		// Scroll with easing (slow start, fast middle, slow end)
		progress := scrolled / totalDistance
		var steps int
		if progress < 0.2 || progress > 0.8 {
			steps = 8 + rand.Intn(5) // Slower at edges
		} else {
			steps = 3 + rand.Intn(3) // Faster in middle
		}

		sb.page.Mouse.Scroll(0, scrollAmount*float64(direction), steps)

		scrolled += scrollAmount

		// Variable pause
		if rand.Float64() < 0.2 {
			// Longer pause (reading)
			time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)
		}

		// Occasional reverse scroll (re-reading)
		if rand.Float64() < 0.1 && scrolled > 100 {
			reverseAmount := 30.0 + rand.Float64()*70
			sb.page.Mouse.Scroll(0, -reverseAmount*float64(direction), 5)
			time.Sleep(time.Duration(300+rand.Intn(500)) * time.Millisecond)
		}
	}

	return nil
}

// ApplyAllBehaviorPatterns applies all human behavior patterns
func ApplyAllBehaviorPatterns(page *rod.Page) error {
	script := `
		// Override mouse event properties to look natural
		const originalDispatchEvent = EventTarget.prototype.dispatchEvent;
		EventTarget.prototype.dispatchEvent = function(event) {
			if (event instanceof MouseEvent) {
				// Add realistic properties
				Object.defineProperty(event, 'isTrusted', {
					get: () => true
				});
			}
			return originalDispatchEvent.call(this, event);
		};
		
		// Make events look more natural
		const originalAddEventListener = EventTarget.prototype.addEventListener;
		EventTarget.prototype.addEventListener = function(type, listener, options) {
			// Wrap listener to add timing variation
			const wrappedListener = function(event) {
				// Small random delay to simulate human reaction time
				setTimeout(() => listener.call(this, event), Math.random() * 10);
			};
			return originalAddEventListener.call(this, type, wrappedListener, options);
		};
		
		// Realistic focus/blur behavior
		let lastActiveTime = Date.now();
		document.addEventListener('visibilitychange', () => {
			lastActiveTime = Date.now();
		});
		
		// Add micro-movements to mouse (humans can't hold perfectly still)
		let mouseX = 0, mouseY = 0;
		document.addEventListener('mousemove', (e) => {
			mouseX = e.clientX;
			mouseY = e.clientY;
		});
	`

	_, err := page.Eval(script)
	return err
}

// TimingJitter adds randomized timing to actions
type TimingJitter struct {
	baseDelay time.Duration
	jitter    float64
}

// NewTimingJitter creates a timing jitter helper
func NewTimingJitter(baseDelay time.Duration, jitterPercent float64) *TimingJitter {
	return &TimingJitter{
		baseDelay: baseDelay,
		jitter:    jitterPercent,
	}
}

// Wait waits with jitter applied
func (tj *TimingJitter) Wait() {
	jitterAmount := float64(tj.baseDelay) * tj.jitter
	actualDelay := float64(tj.baseDelay) + (rand.Float64()*2-1)*jitterAmount
	time.Sleep(time.Duration(actualDelay))
}

// WaitBetweenActions provides human-like timing between actions
func WaitBetweenActions(actionType string) {
	var baseDelay time.Duration

	switch actionType {
	case "click":
		baseDelay = 500 * time.Millisecond
	case "type":
		baseDelay = 200 * time.Millisecond
	case "scroll":
		baseDelay = 300 * time.Millisecond
	case "navigate":
		baseDelay = 1500 * time.Millisecond
	case "read":
		baseDelay = 3 * time.Second
	default:
		baseDelay = 500 * time.Millisecond
	}

	// Add 30% jitter
	jitter := float64(baseDelay) * 0.3
	actualDelay := float64(baseDelay) + (rand.Float64()*2-1)*jitter

	// Occasional longer "distraction" pause (5% chance)
	if rand.Float64() < 0.05 {
		actualDelay += float64(2+rand.Intn(4)) * float64(time.Second)
	}

	time.Sleep(time.Duration(actualDelay))
}

// SimulateUserSession simulates a complete realistic user session
func SimulateUserSession(page *rod.Page, duration time.Duration) error {
	mouse := NewMouseBehavior(page)
	scroll := NewScrollBehavior(page)

	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		// Random action selection
		action := rand.Intn(100)

		switch {
		case action < 40:
			// Scroll (40% chance)
			scrollAmount := 200.0 + rand.Float64()*400
			direction := 1
			if rand.Float64() < 0.2 {
				direction = -1
			}
			scroll.HumanLikeScroll(scrollAmount, direction)

		case action < 60:
			// Random mouse movement (20% chance)
			mouse.RandomMouseMovement(time.Duration(1+rand.Intn(3)) * time.Second)

		case action < 80:
			// Pause and "read" (20% chance)
			time.Sleep(time.Duration(2+rand.Intn(5)) * time.Second)

		default:
			// Idle with micro-movements (20% chance)
			time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
		}
	}

	return nil
}

// InjectBehaviorScripts injects behavior simulation scripts into page
func InjectBehaviorScripts(page *rod.Page) error {
	script := `
		// Create realistic mouse micro-movements
		let realMouseX = 0, realMouseY = 0;
		
		document.addEventListener('mousemove', (e) => {
			realMouseX = e.clientX;
			realMouseY = e.clientY;
		}, { passive: true });
		
		// Simulate natural scroll momentum
		let scrollVelocity = 0;
		let lastScrollTime = 0;
		
		document.addEventListener('wheel', (e) => {
			const now = Date.now();
			const timeDelta = now - lastScrollTime;
			
			if (timeDelta < 100) {
				scrollVelocity = Math.min(scrollVelocity + Math.abs(e.deltaY) * 0.1, 50);
			} else {
				scrollVelocity = Math.abs(e.deltaY) * 0.5;
			}
			
			lastScrollTime = now;
		}, { passive: true });
		
		// Track page interaction patterns
		window.__interactionHistory = [];
		
		['click', 'scroll', 'keypress', 'mousemove'].forEach(eventType => {
			document.addEventListener(eventType, () => {
				window.__interactionHistory.push({
					type: eventType,
					time: Date.now()
				});
				
				// Keep only last 100 interactions
				if (window.__interactionHistory.length > 100) {
					window.__interactionHistory.shift();
				}
			}, { passive: true });
		});
		
		// Natural idle detection
		let lastActivity = Date.now();
		let idleCallbacks = [];
		
		window.onIdle = (callback, threshold = 5000) => {
			idleCallbacks.push({ callback, threshold });
		};
		
		['mousemove', 'keypress', 'scroll', 'click'].forEach(event => {
			document.addEventListener(event, () => {
				lastActivity = Date.now();
			}, { passive: true });
		});
		
		setInterval(() => {
			const idleTime = Date.now() - lastActivity;
			idleCallbacks.forEach(({ callback, threshold }) => {
				if (idleTime >= threshold) {
					callback(idleTime);
				}
			});
		}, 1000);
	`

	_, err := page.Eval(script)
	return err
}

// GetInteractionScore calculates how "human-like" the current session appears
func GetInteractionScore(page *rod.Page) (float64, error) {
	result, err := page.Eval(`
		(() => {
			const history = window.__interactionHistory || [];
			if (history.length < 10) return 0.5;
			
			let score = 1.0;
			
			// Check timing variance
			const timings = [];
			for (let i = 1; i < history.length; i++) {
				timings.push(history[i].time - history[i-1].time);
			}
			
			const avgTiming = timings.reduce((a, b) => a + b, 0) / timings.length;
			const variance = timings.reduce((a, b) => a + Math.pow(b - avgTiming, 2), 0) / timings.length;
			
			// Low variance = bot-like
			if (variance < 1000) score -= 0.3;
			
			// Check event type distribution
			const eventTypes = history.map(h => h.type);
			const uniqueTypes = [...new Set(eventTypes)];
			
			// Only one type = bot-like
			if (uniqueTypes.length <= 1) score -= 0.3;
			
			// Check for realistic patterns
			const hasMouseMove = eventTypes.includes('mousemove');
			const hasScroll = eventTypes.includes('scroll');
			const hasClick = eventTypes.includes('click');
			
			if (!hasMouseMove) score -= 0.1;
			if (!hasScroll) score -= 0.1;
			
			return Math.max(0, Math.min(1, score));
		})()
	`)

	if err != nil {
		return 0, err
	}

	if score, ok := result.Value.Val().(float64); ok {
		return score, nil
	}

	return 0.5, fmt.Errorf("could not parse interaction score")
}
