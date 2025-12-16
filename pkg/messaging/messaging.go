package messaging

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/saurabhkuntal/subspace/pkg/search"
	"github.com/saurabhkuntal/subspace/pkg/stealth"
	"github.com/saurabhkuntal/subspace/pkg/storage"
	"github.com/sirupsen/logrus"
)

// Manager handles messaging operations
type Manager struct {
	stealthMgr *stealth.Manager
	db         *storage.DB
	logger     *logrus.Logger
	maxPerDay  int
}

// NewManager creates a new messaging manager
func NewManager(stealthMgr *stealth.Manager, db *storage.DB, maxPerDay int, logger *logrus.Logger) *Manager {
	return &Manager{
		stealthMgr: stealthMgr,
		db:         db,
		logger:     logger,
		maxPerDay:  maxPerDay,
	}
}

// SendMessage sends a message to a connection
func (m *Manager) SendMessage(page *rod.Page, profileURL, profileName, message string) error {
	m.logger.Infof("Sending message to %s (%s)", profileName, profileURL)

	// Check if we've already sent a message
	alreadySent, err := m.db.HasSentMessage(profileURL)
	if err != nil {
		return fmt.Errorf("failed to check if message already sent: %w", err)
	}
	if alreadySent {
		m.logger.Infof("Already sent message to %s, skipping", profileName)
		return nil
	}

	// Check daily limit
	countToday, err := m.db.GetMessageCountToday()
	if err != nil {
		return fmt.Errorf("failed to get today's message count: %w", err)
	}
	if countToday >= m.maxPerDay {
		return fmt.Errorf("reached daily limit of %d messages", m.maxPerDay)
	}

	// Navigate to messaging page
	messagingURL := fmt.Sprintf("https://www.linkedin.com/messaging/thread/new/?recipients=%s", extractProfileID(profileURL))
	if err := page.Navigate(messagingURL); err != nil {
		// Try alternative: navigate to profile and click Message button
		if err := m.sendMessageFromProfile(page, profileURL, message); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		return nil
	}

	m.stealthMgr.HumanDelay()

	// Find message input
	messageInput, err := page.Timeout(10 * time.Second).Element(".msg-form__contenteditable")
	if err != nil {
		return fmt.Errorf("failed to find message input: %w", err)
	}

	// Type message
	m.logger.Debug("Typing message...")
	if err := m.stealthMgr.HumanType(page, messageInput, message); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	m.stealthMgr.HumanDelay()

	// Find and click Send button
	sendButton, err := page.Timeout(5 * time.Second).Element("button[type='submit'].msg-form__send-button")
	if err != nil {
		return fmt.Errorf("failed to find send button: %w", err)
	}

	m.logger.Debug("Clicking Send button...")
	if err := m.stealthMgr.HumanClick(page, sendButton); err != nil {
		return fmt.Errorf("failed to click send button: %w", err)
	}

	// Wait for confirmation
	time.Sleep(2 * time.Second)

	// Save to database
	msg := &storage.Message{
		ProfileURL:  profileURL,
		ProfileName: profileName,
		Content:     message,
		SentAt:      time.Now(),
	}

	if err := m.db.SaveMessage(msg); err != nil {
		m.logger.Errorf("Failed to save message: %v", err)
	}

	m.logger.Infof("Successfully sent message to %s", profileName)
	
	// Random wait between actions
	m.stealthMgr.RandomWaitBetweenActions()

	return nil
}

// sendMessageFromProfile sends a message by navigating to profile and clicking Message button
func (m *Manager) sendMessageFromProfile(page *rod.Page, profileURL, message string) error {
	// Navigate to profile
	if err := page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	m.stealthMgr.HumanDelay()

	// Find Message button
	messageButton, err := m.findMessageButton(page)
	if err != nil {
		return fmt.Errorf("failed to find message button: %w", err)
	}

	// Click Message button
	if err := m.stealthMgr.HumanClick(page, messageButton); err != nil {
		return fmt.Errorf("failed to click message button: %w", err)
	}

	m.stealthMgr.HumanDelay()

	// Find message input in modal/popup
	messageInput, err := page.Timeout(10 * time.Second).Element(".msg-form__contenteditable")
	if err != nil {
		return fmt.Errorf("failed to find message input: %w", err)
	}

	// Type message
	if err := m.stealthMgr.HumanType(page, messageInput, message); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	m.stealthMgr.HumanDelay()

	// Find and click Send button
	sendButton, err := page.Timeout(5 * time.Second).Element("button[type='submit'].msg-form__send-button")
	if err != nil {
		return fmt.Errorf("failed to find send button: %w", err)
	}

	if err := m.stealthMgr.HumanClick(page, sendButton); err != nil {
		return fmt.Errorf("failed to click send button: %w", err)
	}

	time.Sleep(2 * time.Second)

	return nil
}

// findMessageButton finds the Message button on a profile
func (m *Manager) findMessageButton(page *rod.Page) (*rod.Element, error) {
	selectors := []string{
		"button[aria-label*='Message']",
		"button:has-text('Message')",
		".pvs-profile-actions button:has-text('Message')",
	}

	for _, selector := range selectors {
		button, err := page.Timeout(5 * time.Second).Element(selector)
		if err == nil {
			return button, nil
		}
	}

	return nil, fmt.Errorf("message button not found")
}

// PersonalizeMessage replaces template variables with profile data
func (m *Manager) PersonalizeMessage(template string, profile *search.SearchResult) string {
	message := template

	// Extract first and last name
	nameParts := strings.Fields(profile.Name)
	firstName := ""
	lastName := ""
	if len(nameParts) > 0 {
		firstName = nameParts[0]
	}
	if len(nameParts) > 1 {
		lastName = nameParts[len(nameParts)-1]
	}

	// Replace variables
	message = strings.ReplaceAll(message, "{firstName}", firstName)
	message = strings.ReplaceAll(message, "{lastName}", lastName)
	message = strings.ReplaceAll(message, "{name}", profile.Name)
	message = strings.ReplaceAll(message, "{title}", profile.Title)
	message = strings.ReplaceAll(message, "{company}", profile.Company)

	return message
}

// SendBulkMessages sends messages to multiple connections
func (m *Manager) SendBulkMessages(page *rod.Page, recipients []struct {
	ProfileURL  string
	ProfileName string
	Message     string
}) (int, error) {
	m.logger.Infof("Sending messages to %d recipients...", len(recipients))

	successCount := 0

	for i, recipient := range recipients {
		m.logger.Infof("Processing recipient %d/%d: %s", i+1, len(recipients), recipient.ProfileName)

		// Check daily limit
		countToday, _ := m.db.GetMessageCountToday()
		if countToday >= m.maxPerDay {
			m.logger.Warn("Reached daily limit, stopping")
			break
		}

		// Send message
		err := m.SendMessage(page, recipient.ProfileURL, recipient.ProfileName, recipient.Message)
		if err != nil {
			m.logger.Errorf("Failed to send message: %v", err)
			continue
		}

		successCount++
	}

	m.logger.Infof("Successfully sent %d messages", successCount)
	return successCount, nil
}

// DetectNewConnections detects newly accepted connection requests
func (m *Manager) DetectNewConnections(page *rod.Page) ([]*storage.ConnectionRequest, error) {
	m.logger.Info("Detecting new connections...")

	// Navigate to My Network page
	if err := page.Navigate("https://www.linkedin.com/mynetwork/"); err != nil {
		return nil, fmt.Errorf("failed to navigate to My Network: %w", err)
	}

	m.stealthMgr.HumanDelay()

	// Get pending connection requests from database
	pending, err := m.db.GetPendingConnectionRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}

	var newConnections []*storage.ConnectionRequest

	// Check each pending request
	for _, request := range pending {
		// Navigate to profile
		if err := page.Navigate(request.ProfileURL); err != nil {
			m.logger.Warnf("Failed to navigate to profile %s: %v", request.ProfileURL, err)
			continue
		}

		time.Sleep(2 * time.Second)

		// Check if Connect button is no longer present (meaning they're now connected)
		_, err := page.Timeout(3 * time.Second).Element("button[aria-label*='Connect']")
		if err != nil {
			// Connect button not found, likely now connected
			m.logger.Infof("Detected new connection: %s", request.ProfileName)
			
			// Update status
			m.db.UpdateConnectionRequestStatus(request.ProfileURL, "accepted")
			newConnections = append(newConnections, request)
		}
	}

	m.logger.Infof("Found %d new connections", len(newConnections))
	return newConnections, nil
}

// extractProfileID extracts the profile ID from a LinkedIn URL
func extractProfileID(profileURL string) string {
	// Extract profile ID from URL like https://www.linkedin.com/in/username/
	parts := strings.Split(profileURL, "/in/")
	if len(parts) < 2 {
		return ""
	}
	
	username := strings.TrimSuffix(parts[1], "/")
	return username
}
