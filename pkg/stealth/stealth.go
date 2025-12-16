package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// Manager handles all stealth and anti-detection mechanisms
type Manager struct {
	minDelay               int
	maxDelay               int
	enableRandomScroll     bool
	enableTypingSimulation bool
	enableMouseHover       bool
	operatingHoursStart    int
	operatingHoursEnd      int
	logger                 *logrus.Logger
	rand                   *rand.Rand
}

// NewManager creates a new stealth manager
func NewManager(minDelay, maxDelay int, enableScroll, enableTyping, enableHover bool, startHour, endHour int, logger *logrus.Logger) *Manager {
	return &Manager{
		minDelay:               minDelay,
		maxDelay:               maxDelay,
		enableRandomScroll:     enableScroll,
		enableTypingSimulation: enableTyping,
		enableMouseHover:       enableHover,
		operatingHoursStart:    startHour,
		operatingHoursEnd:      endHour,
		logger:                 logger,
		rand:                   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Point represents a 2D coordinate
type Point struct {
	X, Y float64
}

// RandomDelay introduces a random delay between actions
func (m *Manager) RandomDelay() {
	delay := m.minDelay + m.rand.Intn(m.maxDelay-m.minDelay+1)
	m.logger.Debugf("Waiting for %dms", delay)
	time.Sleep(time.Duration(delay) * time.Millisecond)
}

// HumanDelay simulates human thinking time with variable duration
func (m *Manager) HumanDelay() {
	// Human thinking time: 1-4 seconds with occasional longer pauses
	baseDelay := 1000 + m.rand.Intn(3000)

	// 10% chance of a longer pause (distraction, reading)
	if m.rand.Float64() < 0.1 {
		baseDelay += 3000 + m.rand.Intn(5000)
	}

	m.logger.Debugf("Human thinking delay: %dms", baseDelay)
	time.Sleep(time.Duration(baseDelay) * time.Millisecond)
}

// BezierCurve generates points along a cubic Bezier curve
func (m *Manager) BezierCurve(start, end Point, controlPoints int) []Point {
	points := make([]Point, 0, 50)

	// Generate random control points for natural curve
	ctrl1 := Point{
		X: start.X + (end.X-start.X)*m.rand.Float64(),
		Y: start.Y + (end.Y-start.Y)*m.rand.Float64(),
	}
	ctrl2 := Point{
		X: start.X + (end.X-start.X)*m.rand.Float64(),
		Y: start.Y + (end.Y-start.Y)*m.rand.Float64(),
	}

	// Generate points along the curve
	steps := 50
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		point := m.cubicBezier(start, ctrl1, ctrl2, end, t)
		points = append(points, point)
	}

	return points
}

// cubicBezier calculates a point on a cubic Bezier curve
func (m *Manager) cubicBezier(p0, p1, p2, p3 Point, t float64) Point {
	u := 1 - t
	tt := t * t
	uu := u * u
	uuu := uu * u
	ttt := tt * t

	return Point{
		X: uuu*p0.X + 3*uu*t*p1.X + 3*u*tt*p2.X + ttt*p3.X,
		Y: uuu*p0.Y + 3*uu*t*p1.Y + 3*u*tt*p2.Y + ttt*p3.Y,
	}
}

// HumanMouseMove moves the mouse like a human using Bezier curves with variable speed
func (m *Manager) HumanMouseMove(page *rod.Page, element *rod.Element) error {
	// Get current mouse position (or use random start position)
	startX := float64(m.rand.Intn(100) + 50)
	startY := float64(m.rand.Intn(100) + 50)

	// Get element position
	shape, err := element.Shape()
	if err != nil {
		return err
	}
	box := shape.Box()
	if box == nil {
		return fmt.Errorf("element box is nil")
	}

	// Target position (with some randomness within the element)
	endX := box.X + box.Width/2 + (m.rand.Float64()-0.5)*box.Width*0.3
	endY := box.Y + box.Height/2 + (m.rand.Float64()-0.5)*box.Height*0.3

	start := Point{X: startX, Y: startY}
	end := Point{X: endX, Y: endY}

	// Generate Bezier curve path
	path := m.BezierCurve(start, end, 2)

	// Move along the path with variable speed
	for i, point := range path {
		// Variable speed: slower at start and end, faster in the middle
		var delay time.Duration
		progress := float64(i) / float64(len(path))

		if progress < 0.2 || progress > 0.8 {
			// Slower at start and end
			delay = time.Duration(15+m.rand.Intn(10)) * time.Millisecond
		} else {
			// Faster in the middle
			delay = time.Duration(5+m.rand.Intn(5)) * time.Millisecond
		}

		// Move mouse to point
		err := page.Mouse.MoveLinear(proto.Point{X: point.X, Y: point.Y}, 1)
		if err != nil {
			return err
		}

		time.Sleep(delay)

		// Occasionally add micro-corrections (overshoot and correct)
		if m.rand.Float64() < 0.1 && i < len(path)-5 {
			overshoot := Point{
				X: point.X + (m.rand.Float64()-0.5)*5,
				Y: point.Y + (m.rand.Float64()-0.5)*5,
			}
			page.Mouse.MoveLinear(proto.Point{X: overshoot.X, Y: overshoot.Y}, 1)
			time.Sleep(time.Duration(20+m.rand.Intn(30)) * time.Millisecond)
		}
	}

	// Hover over element briefly before clicking
	if m.enableMouseHover {
		time.Sleep(time.Duration(100+m.rand.Intn(200)) * time.Millisecond)
	}

	return nil
}

// HumanClick performs a human-like click with realistic timing
func (m *Manager) HumanClick(page *rod.Page, element *rod.Element) error {
	// Move mouse to element
	if err := m.HumanMouseMove(page, element); err != nil {
		return err
	}

	// Click duration varies (press time)
	pressDuration := time.Duration(50+m.rand.Intn(100)) * time.Millisecond

	// Click the element
	err := element.Click(proto.InputMouseButtonLeft, 1)
	if err != nil {
		return err
	}

	time.Sleep(pressDuration)

	// Small delay after click
	time.Sleep(time.Duration(50+m.rand.Intn(150)) * time.Millisecond)

	return nil
}

// HumanType simulates realistic human typing with variations
func (m *Manager) HumanType(page *rod.Page, element *rod.Element, text string) error {
	if !m.enableTypingSimulation {
		// Fast typing without simulation
		return element.Input(text)
	}

	// Click on element first
	if err := m.HumanClick(page, element); err != nil {
		return err
	}

	// Type character by character with realistic timing
	for i, char := range text {
		// Base typing speed: 150-300ms per character
		baseDelay := 150 + m.rand.Intn(150)

		// Occasionally type slower (thinking, checking)
		if m.rand.Float64() < 0.1 {
			baseDelay += 200 + m.rand.Intn(500)
		}

		// Type character
		err := element.Input(string(char))
		if err != nil {
			return err
		}

		// Occasionally introduce typos and corrections (5% chance)
		if m.rand.Float64() < 0.05 && i < len(text)-1 {
			// Type wrong character
			wrongChar := string(rune('a' + m.rand.Intn(26)))
			element.Input(wrongChar)
			time.Sleep(time.Duration(100+m.rand.Intn(200)) * time.Millisecond)

			// Backspace
			page.Keyboard.Press(input.Backspace)
			time.Sleep(time.Duration(100+m.rand.Intn(100)) * time.Millisecond)
		}

		time.Sleep(time.Duration(baseDelay) * time.Millisecond)
	}

	return nil
}

// RandomScroll performs random scrolling behavior on the page
func (m *Manager) RandomScroll(page *rod.Page) error {
	if !m.enableRandomScroll {
		return nil
	}

	// Number of scroll actions
	scrolls := 1 + m.rand.Intn(4)

	for i := 0; i < scrolls; i++ {
		// Random scroll distance (can be negative to scroll back up)
		scrollDistance := float64(100 + m.rand.Intn(400))
		if m.rand.Float64() < 0.2 {
			// 20% chance to scroll back up
			scrollDistance = -scrollDistance
		}

		// Smooth scroll with variable speed
		steps := 5 + m.rand.Intn(10)
		stepDistance := scrollDistance / float64(steps)

		for j := 0; j < steps; j++ {
			err := page.Mouse.Scroll(0, stepDistance, 1)
			if err != nil {
				return err
			}

			// Variable delay between scroll steps
			delay := 20 + m.rand.Intn(50)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}

		// Pause after scrolling
		time.Sleep(time.Duration(200+m.rand.Intn(800)) * time.Millisecond)
	}

	return nil
}

// RandomHover randomly hovers over elements on the page
func (m *Manager) RandomHover(page *rod.Page, elements rod.Elements) error {
	if !m.enableMouseHover || len(elements) == 0 {
		return nil
	}

	// Select random element to hover
	element := elements[m.rand.Intn(len(elements))]

	// Move mouse to element
	shape, err := element.Shape()
	if err != nil {
		return err
	}
	box := shape.Box()

	x := box.X + box.Width/2 + (m.rand.Float64()-0.5)*box.Width*0.3
	y := box.Y + box.Height/2 + (m.rand.Float64()-0.5)*box.Height*0.3

	err = page.Mouse.MoveLinear(proto.Point{X: x, Y: y}, 10)
	if err != nil {
		return err
	}

	// Hover duration
	time.Sleep(time.Duration(300+m.rand.Intn(700)) * time.Millisecond)

	return nil
}

// IsWithinOperatingHours checks if current time is within configured operating hours
func (m *Manager) IsWithinOperatingHours() bool {
	currentHour := time.Now().Hour()

	if m.operatingHoursStart <= m.operatingHoursEnd {
		// Normal range (e.g., 9-17)
		return currentHour >= m.operatingHoursStart && currentHour < m.operatingHoursEnd
	}

	// Overnight range (e.g., 22-6)
	return currentHour >= m.operatingHoursStart || currentHour < m.operatingHoursEnd
}

// WaitForOperatingHours waits until we're within operating hours
func (m *Manager) WaitForOperatingHours() {
	for !m.IsWithinOperatingHours() {
		m.logger.Info("Outside operating hours, waiting...")
		time.Sleep(15 * time.Minute)
	}
}

// SimulateBreak simulates a random break period (coffee, lunch, etc.)
func (m *Manager) SimulateBreak() {
	// 5% chance of taking a break after any action
	if m.rand.Float64() < 0.05 {
		breakDuration := time.Duration(3+m.rand.Intn(10)) * time.Minute
		m.logger.Infof("Taking a break for %v", breakDuration)
		time.Sleep(breakDuration)
	}
}

// RandomWaitBetweenActions combines multiple delays and checks
func (m *Manager) RandomWaitBetweenActions() {
	m.RandomDelay()
	m.SimulateBreak()
	// Operating hours check removed - run anytime
}
