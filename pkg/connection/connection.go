package connection

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/saurabhkuntal/subspace/pkg/search"
	"github.com/saurabhkuntal/subspace/pkg/stealth"
	"github.com/saurabhkuntal/subspace/pkg/storage"
	"github.com/sirupsen/logrus"
)

// Manager handles connection request operations
type Manager struct {
	stealthMgr      *stealth.Manager
	db              *storage.DB
	logger          *logrus.Logger
	maxPerDay       int
	maxPerHour      int
	messageTemplate string
}

// NewManager creates a new connection manager
func NewManager(stealthMgr *stealth.Manager, db *storage.DB, maxPerDay, maxPerHour int, messageTemplate string, logger *logrus.Logger) *Manager {
	return &Manager{
		stealthMgr:      stealthMgr,
		db:              db,
		logger:          logger,
		maxPerDay:       maxPerDay,
		maxPerHour:      maxPerHour,
		messageTemplate: messageTemplate,
	}
}

// SendConnectionRequest sends a connection request to a profile
func (m *Manager) SendConnectionRequest(page *rod.Page, profile *search.SearchResult, customMessage string) error {
	m.logger.Infof("Sending connection request to %s (%s)", profile.Name, profile.ProfileURL)

	// Check if we've already sent a request to this profile
	alreadySent, err := m.db.HasSentConnectionRequest(profile.ProfileURL)
	if err != nil {
		return fmt.Errorf("failed to check if request already sent: %w", err)
	}
	if alreadySent {
		m.logger.Infof("Already sent connection request to %s, skipping", profile.Name)
		return nil
	}

	// Check daily limit
	countToday, err := m.db.GetConnectionRequestCountToday()
	if err != nil {
		return fmt.Errorf("failed to get today's connection count: %w", err)
	}
	if countToday >= m.maxPerDay {
		return fmt.Errorf("reached daily limit of %d connection requests", m.maxPerDay)
	}

	// Check hourly limit
	countThisHour, err := m.db.GetConnectionRequestCountThisHour()
	if err != nil {
		return fmt.Errorf("failed to get this hour's connection count: %w", err)
	}
	if countThisHour >= m.maxPerHour {
		return fmt.Errorf("reached hourly limit of %d connection requests", m.maxPerHour)
	}

	// Navigate to profile
	if err := page.Navigate(profile.ProfileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	// Wait for page to load
	time.Sleep(3 * time.Second)

	m.stealthMgr.HumanDelay()

	// Random scroll on profile to appear natural
	if err := m.stealthMgr.RandomScroll(page); err != nil {
		m.logger.Warnf("Failed to scroll: %v", err)
	}

	m.stealthMgr.RandomDelay()

	// Find Connect button (may be direct or in More dropdown)
	connectButton, err := m.findConnectButton(page)
	if err != nil {
		return fmt.Errorf("failed to find connect button: %w", err)
	}

	// If connectButton is not nil, we need to click it
	// If it's nil, the button was already clicked via JavaScript in findConnectButton
	if connectButton != nil {
		m.logger.Debug("Clicking Connect button...")
		if err := m.stealthMgr.HumanClick(page, connectButton); err != nil {
			// Try direct click if human click fails
			if err := connectButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
				return fmt.Errorf("failed to click connect button: %w", err)
			}
		}
	}

	m.stealthMgr.HumanDelay()
	time.Sleep(1 * time.Second)

	// Handle the connection modal
	// LinkedIn shows "Add a note to your invitation?" modal with two options:
	// 1. "Add a note" button
	// 2. "Send without a note" button

	if customMessage != "" {
		// Try to add a personalized note
		m.logger.Debug("Attempting to add personalized note...")
		if err := m.handleConnectionModal(page, profile, customMessage); err != nil {
			m.logger.Warnf("Failed to add note, trying to send without: %v", err)
			// Fall back to sending without a note
			if err := m.clickSendWithoutNote(page); err != nil {
				return fmt.Errorf("failed to send connection request: %w", err)
			}
		}
	} else {
		// Send without a note
		m.logger.Debug("Sending connection request without note...")
		if err := m.clickSendWithoutNote(page); err != nil {
			return fmt.Errorf("failed to send connection request: %w", err)
		}
	}

	// Wait for confirmation
	time.Sleep(2 * time.Second)

	// Save to database
	request := &storage.ConnectionRequest{
		ProfileURL:  profile.ProfileURL,
		ProfileName: profile.Name,
		Company:     profile.Company,
		Message:     customMessage,
		SentAt:      time.Now(),
		Status:      "pending",
	}

	if err := m.db.SaveConnectionRequest(request); err != nil {
		m.logger.Errorf("Failed to save connection request: %v", err)
	}

	m.logger.Infof("Successfully sent connection request to %s", profile.Name)

	// Random wait between actions
	m.stealthMgr.RandomWaitBetweenActions()

	return nil
}

// handleConnectionModal handles the "Add a note to your invitation?" modal
func (m *Manager) handleConnectionModal(page *rod.Page, profile *search.SearchResult, customMessage string) error {
	// Click "Add a note" button
	addNoteSelectors := []string{
		"button[aria-label='Add a note']",
		"button:has-text('Add a note')",
		".artdeco-modal button:has-text('Add a note')",
	}

	var addNoteBtn *rod.Element
	for _, selector := range addNoteSelectors {
		btn, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil && btn != nil {
			addNoteBtn = btn
			break
		}
	}

	if addNoteBtn == nil {
		return fmt.Errorf("add a note button not found")
	}

	if err := addNoteBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click add a note: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Find note textarea
	textareaSelectors := []string{
		"textarea[name='message']",
		"textarea#custom-message",
		".artdeco-modal textarea",
		"textarea",
	}

	var noteTextarea *rod.Element
	for _, selector := range textareaSelectors {
		ta, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil && ta != nil {
			noteTextarea = ta
			break
		}
	}

	if noteTextarea == nil {
		return fmt.Errorf("note textarea not found")
	}

	// Process message template
	personalizedMessage := m.personalizeMessage(customMessage, profile)

	// Ensure message is within LinkedIn's character limit (300 characters)
	if len(personalizedMessage) > 300 {
		personalizedMessage = personalizedMessage[:297] + "..."
	}

	// Type the message
	if err := m.stealthMgr.HumanType(page, noteTextarea, personalizedMessage); err != nil {
		// Fallback to direct input
		noteTextarea.Input(personalizedMessage)
	}

	time.Sleep(500 * time.Millisecond)

	// Click Send button
	sendSelectors := []string{
		"button[aria-label='Send invitation']",
		"button[aria-label='Send now']",
		"button[aria-label='Send']",
		".artdeco-modal button:has-text('Send')",
		"button.artdeco-button--primary:has-text('Send')",
	}

	for _, selector := range sendSelectors {
		sendBtn, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil && sendBtn != nil {
			if err := sendBtn.Click(proto.InputMouseButtonLeft, 1); err == nil {
				m.logger.Debug("Clicked Send button")
				return nil
			}
		}
	}

	return fmt.Errorf("send button not found")
}

// clickSendWithoutNote clicks the "Send without a note" button
func (m *Manager) clickSendWithoutNote(page *rod.Page) error {
	// Wait for modal to appear
	time.Sleep(1 * time.Second)

	// Take screenshot for debugging
	page.MustScreenshot("screenshots/connection_modal.png")
	m.logger.Debug("Screenshot saved: screenshots/connection_modal.png")

	// Selectors for "Send without a note" button - try multiple variations
	sendWithoutNoteSelectors := []string{
		"button[aria-label='Send without a note']",
		"button[aria-label='Send now']",
		"button[aria-label='Send invitation']",
		"button:has-text('Send without a note')",
		"button:has-text('Send now')",
		".artdeco-modal button:has-text('Send without a note')",
		".artdeco-modal button:has-text('Send')",
		"[role='dialog'] button:has-text('Send')",
		"button.artdeco-button--primary",
	}

	for _, selector := range sendWithoutNoteSelectors {
		btn, err := page.Timeout(2 * time.Second).Element(selector)
		if err == nil && btn != nil {
			visible, _ := btn.Visible()
			if visible {
				text, _ := btn.Text()
				m.logger.Debugf("Found button with selector %s, text: %s", selector, text)
				// Make sure it's a send-related button
				textLower := strings.ToLower(text)
				if strings.Contains(textLower, "send") || strings.Contains(textLower, "connect") {
					if err := btn.Click(proto.InputMouseButtonLeft, 1); err == nil {
						m.logger.Debug("Clicked Send button")
						return nil
					}
				}
			}
		}
	}

	// Try JavaScript fallback - more aggressive
	result, err := page.Eval(`() => {
		// Find all buttons in modal or dialog
		const allButtons = document.querySelectorAll('.artdeco-modal button, [role="dialog"] button, .send-invite button, button');
		for (const btn of allButtons) {
			const text = btn.innerText.toLowerCase().trim();
			const ariaLabel = (btn.getAttribute('aria-label') || '').toLowerCase();
			
			// Check for send-related buttons
			if (text.includes('send without') || text === 'send' || text.includes('send now') ||
			    ariaLabel.includes('send without') || ariaLabel.includes('send now') || ariaLabel.includes('send invitation')) {
				console.log('Found send button:', text, ariaLabel);
				btn.click();
				return {success: true, text: text};
			}
		}
		
		// Try finding primary button in modal
		const modal = document.querySelector('.artdeco-modal, [role="dialog"]');
		if (modal) {
			const primaryBtn = modal.querySelector('button.artdeco-button--primary, button.artdeco-button--2');
			if (primaryBtn) {
				console.log('Found primary button:', primaryBtn.innerText);
				primaryBtn.click();
				return {success: true, text: primaryBtn.innerText};
			}
		}
		
		return {success: false, buttons: Array.from(allButtons).slice(0, 10).map(b => b.innerText.trim())};
	}`)

	if err == nil && result.Value.Val() != nil {
		data := result.Value.Val().(map[string]interface{})
		if success, ok := data["success"].(bool); ok && success {
			m.logger.Debugf("Clicked send button via JavaScript: %v", data["text"])
			return nil
		}
		// Log available buttons for debugging
		if buttons, ok := data["buttons"].([]interface{}); ok {
			m.logger.Debugf("Available buttons: %v", buttons)
		}
	}

	return fmt.Errorf("send without a note button not found")
}

// findConnectButton finds the Connect button on a profile
// LinkedIn has multiple layouts:
// 1. Direct "Connect" button visible on profile
// 2. "Connect" hidden under "More" dropdown menu
// 3. "Add" option in the More dropdown (same as Connect)
func (m *Manager) findConnectButton(page *rod.Page) (*rod.Element, error) {
	// Wait for page to load
	time.Sleep(2 * time.Second)

	// First, try direct Connect button selectors
	directSelectors := []string{
		"button[aria-label='Invite to connect']",
		"button[aria-label*='Connect']",
		"main button:has-text('Connect')",
		".pvs-profile-actions button:has-text('Connect')",
		"button.artdeco-button--primary:has-text('Connect')",
	}

	for _, selector := range directSelectors {
		button, err := page.Timeout(2 * time.Second).Element(selector)
		if err == nil && button != nil {
			// Verify button is visible and not hidden
			visible, _ := button.Visible()
			if visible {
				m.logger.Debug("Found direct Connect button")
				return button, nil
			}
		}
	}

	m.logger.Debug("Direct Connect button not found, checking More dropdown...")

	// Try to find and click the "More" button to reveal Connect/Add option
	moreButtonSelectors := []string{
		"button[aria-label='More actions']",
		"button[aria-label='More']",
		"div.pvs-profile-actions button:has-text('More')",
		"main section button:has-text('More')",
		".artdeco-dropdown__trigger:has-text('More')",
	}

	var moreButton *rod.Element
	for _, selector := range moreButtonSelectors {
		btn, err := page.Timeout(2 * time.Second).Element(selector)
		if err == nil && btn != nil {
			visible, _ := btn.Visible()
			if visible {
				moreButton = btn
				m.logger.Debugf("Found More button with selector: %s", selector)
				break
			}
		}
	}

	if moreButton != nil {
		// Click the More button to open dropdown
		m.logger.Debug("Clicking More button to reveal Connect option...")
		if err := moreButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
			m.logger.Warnf("Failed to click More button: %v", err)
		}

		// Wait for dropdown to appear
		time.Sleep(1 * time.Second)

		// Look for Connect or Add in the dropdown
		dropdownSelectors := []string{
			"div[aria-label='Invite to connect']",
			"div.artdeco-dropdown__item:has-text('Connect')",
			"div.artdeco-dropdown__item:has-text('Add')",
			"li.artdeco-dropdown__item:has-text('Connect')",
			"li.artdeco-dropdown__item:has-text('Add')",
			"span:has-text('Connect')",
			"span:has-text('Add')",
		}

		for _, selector := range dropdownSelectors {
			item, err := page.Timeout(2 * time.Second).Element(selector)
			if err == nil && item != nil {
				visible, _ := item.Visible()
				if visible {
					m.logger.Debugf("Found Connect/Add option in dropdown with selector: %s", selector)
					return item, nil
				}
			}
		}

		// Try JavaScript to find the dropdown item
		m.logger.Debug("Using JavaScript to find dropdown Connect/Add option...")
		result, err := page.Eval(`() => {
			// Look for dropdown items
			const items = document.querySelectorAll('.artdeco-dropdown__item, [role="menuitem"]');
			for (const item of items) {
				const text = item.innerText.toLowerCase();
				if (text.includes('connect') || text === 'add') {
					return true;
				}
			}
			return false;
		}`)

		if err == nil && result.Value.Bool() {
			// Find and return the element
			connectItem, err := page.Eval(`() => {
				const items = document.querySelectorAll('.artdeco-dropdown__item, [role="menuitem"]');
				for (const item of items) {
					const text = item.innerText.toLowerCase();
					if (text.includes('connect') || text === 'add') {
						item.click();
						return true;
					}
				}
				return false;
			}`)

			if err == nil && connectItem.Value.Bool() {
				m.logger.Debug("Clicked Connect/Add via JavaScript")
				// Return nil to indicate button was already clicked
				return nil, nil
			}
		}
	}

	return nil, fmt.Errorf("connect button not found")
}

// personalizeMessage replaces template variables with profile data
func (m *Manager) personalizeMessage(template string, profile *search.SearchResult) string {
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

// SendBulkConnectionRequests sends connection requests to multiple profiles
func (m *Manager) SendBulkConnectionRequests(page *rod.Page, profiles []*search.SearchResult, customMessage string) (int, error) {
	m.logger.Infof("Sending connection requests to %d profiles...", len(profiles))

	successCount := 0

	for i, profile := range profiles {
		m.logger.Infof("Processing profile %d/%d: %s", i+1, len(profiles), profile.Name)

		// Check if we've reached limits
		countToday, _ := m.db.GetConnectionRequestCountToday()
		if countToday >= m.maxPerDay {
			m.logger.Warn("Reached daily limit, stopping")
			break
		}

		countThisHour, _ := m.db.GetConnectionRequestCountThisHour()
		if countThisHour >= m.maxPerHour {
			m.logger.Warn("Reached hourly limit, waiting...")
			time.Sleep(time.Hour)
		}

		// Send connection request
		err := m.SendConnectionRequest(page, profile, customMessage)
		if err != nil {
			m.logger.Errorf("Failed to send connection request: %v", err)
			continue
		}

		successCount++
	}

	m.logger.Infof("Successfully sent %d connection requests", successCount)
	return successCount, nil
}

// GetPendingConnections retrieves pending connection requests
func (m *Manager) GetPendingConnections() ([]*storage.ConnectionRequest, error) {
	return m.db.GetPendingConnectionRequests()
}

// UpdateConnectionStatus updates the status of a connection request
func (m *Manager) UpdateConnectionStatus(profileURL, status string) error {
	return m.db.UpdateConnectionRequestStatus(profileURL, status)
}
