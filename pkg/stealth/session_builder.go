package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// SessionBuilder handles gradual account "warming" to build reputation
type SessionBuilder struct {
	logger           *logrus.Logger
	sessionAge       time.Duration
	actionsPerformed int
	dailyActions     map[string]int
	lastActionTime   time.Time
}

// NewSessionBuilder creates a new session builder
func NewSessionBuilder(logger *logrus.Logger) *SessionBuilder {
	return &SessionBuilder{
		logger:       logger,
		dailyActions: make(map[string]int),
	}
}

// WarmUpSession performs natural browsing activities to establish account history
func (sb *SessionBuilder) WarmUpSession(page *rod.Page, duration time.Duration) error {
	sb.logger.Infof("Starting session warm-up for %v...", duration)

	endTime := time.Now().Add(duration)
	activities := []func(*rod.Page) error{
		sb.browseHomeFeed,
		sb.viewNotifications,
		sb.checkMessages,
		sb.viewProfile,
		sb.browseNetwork,
		sb.readArticle,
	}

	for time.Now().Before(endTime) {
		// Random activity selection
		activity := activities[rand.Intn(len(activities))]

		sb.logger.Debug("Performing warm-up activity...")
		if err := activity(page); err != nil {
			sb.logger.Warnf("Warm-up activity failed: %v", err)
		}

		sb.actionsPerformed++
		sb.lastActionTime = time.Now()

		// Random pause between activities (30s - 3min)
		pause := time.Duration(30+rand.Intn(150)) * time.Second
		sb.logger.Debugf("Pausing for %v before next activity...", pause)
		time.Sleep(pause)
	}

	sb.logger.Info("Session warm-up completed")
	return nil
}

// browseHomeFeed simulates browsing the home feed
func (sb *SessionBuilder) browseHomeFeed(page *rod.Page) error {
	sb.logger.Debug("Browsing home feed...")

	if err := page.Navigate("https://www.linkedin.com/feed/"); err != nil {
		return err
	}

	time.Sleep(time.Duration(3+rand.Intn(4)) * time.Second)

	// Scroll through feed
	for i := 0; i < 5+rand.Intn(5); i++ {
		scrollAmount := 300 + rand.Intn(200)
		page.Mouse.Scroll(0, float64(scrollAmount), 5)
		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)

		// Occasionally scroll up (reading behavior)
		if rand.Float64() < 0.3 {
			page.Mouse.Scroll(0, float64(-100-rand.Intn(100)), 3)
			time.Sleep(time.Duration(1+rand.Intn(2)) * time.Second)
		}
	}

	// Randomly interact with a post
	if rand.Float64() < 0.4 {
		sb.interactWithPost(page)
	}

	return nil
}

// interactWithPost simulates liking or commenting on a post
func (sb *SessionBuilder) interactWithPost(page *rod.Page) error {
	sb.logger.Debug("Interacting with post...")

	// Find like buttons
	likeButtons, err := page.Elements("button[aria-label*='Like']")
	if err != nil || len(likeButtons) == 0 {
		return fmt.Errorf("no like buttons found")
	}

	// Click a random like button
	if len(likeButtons) > 0 {
		idx := rand.Intn(len(likeButtons))
		if idx < len(likeButtons) {
			likeButtons[idx].MustClick()
			time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)
		}
	}

	return nil
}

// viewNotifications checks the notifications page
func (sb *SessionBuilder) viewNotifications(page *rod.Page) error {
	sb.logger.Debug("Viewing notifications...")

	if err := page.Navigate("https://www.linkedin.com/notifications/"); err != nil {
		return err
	}

	time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)

	// Scroll through notifications
	for i := 0; i < 3+rand.Intn(3); i++ {
		page.Mouse.Scroll(0, float64(200+rand.Intn(150)), 4)
		time.Sleep(time.Duration(1+rand.Intn(2)) * time.Second)
	}

	return nil
}

// checkMessages views the messaging page
func (sb *SessionBuilder) checkMessages(page *rod.Page) error {
	sb.logger.Debug("Checking messages...")

	if err := page.Navigate("https://www.linkedin.com/messaging/"); err != nil {
		return err
	}

	time.Sleep(time.Duration(3+rand.Intn(3)) * time.Second)

	// Scroll through conversations
	page.Mouse.Scroll(0, float64(200+rand.Intn(100)), 4)
	time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)

	return nil
}

// viewProfile views own profile
func (sb *SessionBuilder) viewProfile(page *rod.Page) error {
	sb.logger.Debug("Viewing own profile...")

	if err := page.Navigate("https://www.linkedin.com/in/me/"); err != nil {
		return err
	}

	time.Sleep(time.Duration(3+rand.Intn(4)) * time.Second)

	// Scroll through profile
	for i := 0; i < 3+rand.Intn(2); i++ {
		page.Mouse.Scroll(0, float64(300+rand.Intn(200)), 5)
		time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)
	}

	return nil
}

// browseNetwork views connections/network page
func (sb *SessionBuilder) browseNetwork(page *rod.Page) error {
	sb.logger.Debug("Browsing network...")

	if err := page.Navigate("https://www.linkedin.com/mynetwork/"); err != nil {
		return err
	}

	time.Sleep(time.Duration(3+rand.Intn(3)) * time.Second)

	// Scroll through network suggestions
	for i := 0; i < 4+rand.Intn(3); i++ {
		page.Mouse.Scroll(0, float64(250+rand.Intn(150)), 4)
		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)
	}

	return nil
}

// readArticle simulates reading a LinkedIn article
func (sb *SessionBuilder) readArticle(page *rod.Page) error {
	sb.logger.Debug("Reading article...")

	// Navigate to feed first to find articles
	if err := page.Navigate("https://www.linkedin.com/feed/"); err != nil {
		return err
	}

	time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)

	// Find article links
	articleLinks, err := page.Elements("a[href*='/pulse/']")
	if err != nil || len(articleLinks) == 0 {
		sb.logger.Debug("No articles found, skipping...")
		return nil
	}

	// Click first article
	if len(articleLinks) > 0 {
		if href, _ := articleLinks[0].Attribute("href"); href != nil {
			page.Navigate(*href)
			time.Sleep(time.Duration(5+rand.Intn(10)) * time.Second)

			// Scroll through article
			for i := 0; i < 3+rand.Intn(4); i++ {
				page.Mouse.Scroll(0, float64(400+rand.Intn(200)), 6)
				time.Sleep(time.Duration(3+rand.Intn(4)) * time.Second)
			}
		}
	}

	return nil
}

// GradualActionIncrease slowly increases daily actions over time
func (sb *SessionBuilder) GradualActionIncrease(currentDay int, baseActions int) int {
	// Start with 20% of normal, increase by 10% per day
	multiplier := 0.2 + (float64(currentDay) * 0.1)
	if multiplier > 1.0 {
		multiplier = 1.0
	}

	actions := int(float64(baseActions) * multiplier)
	sb.logger.Debugf("Day %d: Allowing %d actions (%.0f%% of normal)", currentDay, actions, multiplier*100)

	return actions
}

// RecordAction records an action for rate limiting
func (sb *SessionBuilder) RecordAction(actionType string) {
	today := time.Now().Format("2006-01-02")
	sb.dailyActions[today+":"+actionType]++
	sb.lastActionTime = time.Now()
}

// ShouldAllowAction checks if action should be allowed based on reputation building strategy
func (sb *SessionBuilder) ShouldAllowAction(actionType string, maxPerDay int, accountAge time.Duration) bool {
	today := time.Now().Format("2006-01-02")
	key := today + ":" + actionType
	current := sb.dailyActions[key]

	// Adjust limits based on account age
	daysSinceCreation := int(accountAge.Hours() / 24)
	adjustedMax := sb.GradualActionIncrease(daysSinceCreation, maxPerDay)

	return current < adjustedMax
}

// GetReputationScore calculates a reputation score based on account behavior
func (sb *SessionBuilder) GetReputationScore() float64 {
	score := 0.0

	// Age contributes to reputation
	daysSinceFirstUse := time.Since(sb.lastActionTime).Hours() / 24
	ageScore := daysSinceFirstUse / 30.0 // 30 days = full age score
	if ageScore > 1.0 {
		ageScore = 1.0
	}
	score += ageScore * 0.4

	// Total actions contribute
	if sb.actionsPerformed > 100 {
		score += 0.3
	} else {
		score += (float64(sb.actionsPerformed) / 100.0) * 0.3
	}

	// Daily consistency contributes
	uniqueDays := len(sb.dailyActions)
	consistencyScore := float64(uniqueDays) / 30.0
	if consistencyScore > 1.0 {
		consistencyScore = 1.0
	}
	score += consistencyScore * 0.3

	return score
}

// SimulateHumanSchedule simulates realistic human usage patterns
func (sb *SessionBuilder) SimulateHumanSchedule() time.Duration {
	hour := time.Now().Hour()

	// Active hours: 8am-10pm with peak at 12pm and 6pm
	if hour >= 8 && hour <= 22 {
		// Higher activity during peak hours
		if hour == 12 || hour == 18 {
			return time.Duration(10+rand.Intn(20)) * time.Minute
		}
		// Normal activity
		return time.Duration(20+rand.Intn(40)) * time.Minute
	}

	// Low/no activity at night
	return time.Duration(2+rand.Intn(4)) * time.Hour
}

// RecommendedWarmUpPlan returns a recommended warm-up schedule
func (sb *SessionBuilder) RecommendedWarmUpPlan() string {
	return `
🔥 Recommended Account Warm-Up Plan:

Week 1: Foundation Building
- Day 1-2: Only view feed, read articles (15-20 min/day)
- Day 3-4: Add profile viewing, check notifications (20-25 min/day)
- Day 5-7: Light engagement - like 2-3 posts (25-30 min/day)

Week 2: Gradual Increase
- Day 8-10: View 5-7 profiles manually, like 5-7 posts (30-35 min/day)
- Day 11-14: Send 1-2 connection requests/day (35-40 min/day)

Week 3: Establish Pattern
- Day 15-21: Send 3-5 connection requests/day, send 1-2 messages (40-45 min/day)

Week 4+: Normal Usage
- Day 22+: Full automation with limits (up to 50 actions/day)

💡 Tips:
- Vary your login times each day
- Take 1-2 days off per week
- Mix automated and manual activities
- Never exceed LinkedIn's daily limits
- Monitor for any warnings or restrictions
`
}
