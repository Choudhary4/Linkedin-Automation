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

	// Navigate to profile and use Message button (most reliable method)
	if err := m.sendMessageFromProfile(page, profileURL, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

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

// sendMessageFromProfile sends a message to a 1st-degree connection
func (m *Manager) sendMessageFromProfile(page *rod.Page, profileURL, message string) error {
	// CRITICAL: Close ALL open chat overlays first to avoid sending to wrong person
	m.logger.Debug("Closing ALL open chat overlays...")
	page.Eval(`() => {
		// Find all X/close buttons in chat overlays and click them
		const closeSelectors = [
			'.msg-overlay-bubble-header__control[data-control-name="overlay.close_conversation_window"]',
			'.msg-overlay-bubble-header button[data-control-name="overlay.close_conversation_window"]',
			'button.msg-overlay-bubble-header__control',
			'.msg-overlay-list-bubble__convo-card-container button[aria-label*="Close"]',
		];
		
		for (const selector of closeSelectors) {
			const btns = document.querySelectorAll(selector);
			btns.forEach(btn => {
				try { btn.click(); } catch(e) {}
			});
		}
		
		// Also click X buttons directly
		const allBtns = document.querySelectorAll('.msg-overlay-bubble-header button');
		allBtns.forEach(btn => {
			const svg = btn.querySelector('svg');
			if (svg || btn.getAttribute('data-control-name')?.includes('close')) {
				try { btn.click(); } catch(e) {}
			}
		});
	}`)
	time.Sleep(2 * time.Second)

	// Navigate to the profile page
	m.logger.Infof("Navigating to profile: %s", profileURL)

	if err := page.Navigate(profileURL); err != nil {
		return fmt.Errorf("failed to navigate to profile: %w", err)
	}

	time.Sleep(4 * time.Second) // Wait for page to load

	// Close any chat overlays AGAIN after navigation
	page.Eval(`() => {
		const allBtns = document.querySelectorAll('.msg-overlay-bubble-header button');
		allBtns.forEach(btn => {
			const svg = btn.querySelector('svg');
			const name = btn.getAttribute('data-control-name') || '';
			if (svg || name.includes('close')) {
				try { btn.click(); } catch(e) {}
			}
		});
	}`)
	time.Sleep(1 * time.Second)

	// Click Message button on profile - this opens chat overlay for 1st degree connections
	m.logger.Debug("Looking for Message button on profile...")

	messageClicked := false

	// Method 1: JavaScript - find and click Message button
	result, _ := page.Eval(`() => {
		// Look for Message button in profile actions
		const buttons = document.querySelectorAll('button');
		for (const btn of buttons) {
			const text = btn.textContent.trim();
			if (text === 'Message') {
				btn.click();
				return true;
			}
		}
		// Also try aria-label
		const msgBtn = document.querySelector('button[aria-label*="Message"]');
		if (msgBtn) {
			msgBtn.click();
			return true;
		}
		return false;
	}`)
	if result != nil && result.Value.Bool() {
		messageClicked = true
		m.logger.Info("✅ Clicked Message button via JavaScript")
	}

	// Method 2: Direct element selector
	if !messageClicked {
		msgBtn, err := page.Timeout(5 * time.Second).Element("button[aria-label*='Message']")
		if err == nil && msgBtn != nil {
			if clickErr := msgBtn.Click("left", 1); clickErr == nil {
				messageClicked = true
				m.logger.Info("✅ Clicked Message button via element click")
			}
		}
	}

	if !messageClicked {
		return fmt.Errorf("failed to find Message button on profile")
	}

	// Wait for chat overlay to open
	time.Sleep(3 * time.Second)

	// Extract expected name from profile URL for verification
	expectedUsername := extractProfileID(profileURL)

	// IMPORTANT: Click on the chat header to make sure THIS chat is focused/active
	page.Eval(`() => {
		// Find all chat overlays and click on the last one (most recently opened)
		const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
		if (overlays.length > 0) {
			const lastOverlay = overlays[overlays.length - 1];
			lastOverlay.click();
			// Also click on the contenteditable inside it
			const input = lastOverlay.querySelector('.msg-form__contenteditable, div[contenteditable="true"]');
			if (input) {
				input.focus();
				input.click();
			}
		}
	}`)
	time.Sleep(500 * time.Millisecond)

	// Verify we're chatting with the correct person by checking the chat header
	chatHeaderName, _ := page.Eval(`() => {
		// Get the name from the most recently opened/focused chat overlay
		const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
		if (overlays.length > 0) {
			const lastOverlay = overlays[overlays.length - 1];
			const header = lastOverlay.querySelector('.msg-overlay-bubble-header__title');
			if (header) return header.textContent.trim();
		}
		return '';
	}`)

	if chatHeaderName != nil {
		headerName := chatHeaderName.Value.String()
		m.logger.Infof("Chat opened with: %s (expected profile: %s)", headerName, expectedUsername)
	}

	// Find message input in the LAST (most recently opened) chat overlay
	m.logger.Debug("Looking for message input in chat overlay...")

	// Use JavaScript to find and focus the input in the LAST chat overlay
	page.Eval(`() => {
		const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
		if (overlays.length > 0) {
			const lastOverlay = overlays[overlays.length - 1];
			const input = lastOverlay.querySelector('.msg-form__contenteditable, div[contenteditable="true"]');
			if (input) {
				input.focus();
				input.click();
			}
		}
	}`)
	time.Sleep(500 * time.Millisecond)

	var messageInput *rod.Element
	inputSelectors := []string{
		".msg-form__contenteditable",
		"div[role='textbox'][contenteditable='true']",
	}

	for _, sel := range inputSelectors {
		input, err := page.Timeout(8 * time.Second).Element(sel)
		if err == nil && input != nil {
			messageInput = input
			m.logger.Infof("Found message input with selector: %s", sel)
			break
		}
	}

	if messageInput == nil {
		return fmt.Errorf("failed to find message input in chat overlay")
	}

	// Click to focus the input in the LAST overlay
	messageInput.Click("left", 1)
	time.Sleep(500 * time.Millisecond)

	// Focus using JavaScript - target the LAST overlay specifically
	page.Eval(`() => {
		const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
		if (overlays.length > 0) {
			const lastOverlay = overlays[overlays.length - 1];
			const input = lastOverlay.querySelector('.msg-form__contenteditable, div[contenteditable="true"]');
			if (input) {
				input.focus();
				input.click();
			}
		}
	}`)
	time.Sleep(500 * time.Millisecond)

	// Type the message using keyboard simulation
	m.logger.Infof("Typing message: %s", message)
	page.InsertText(message)
	time.Sleep(1 * time.Second)

	// Verify message was typed in the LAST overlay
	hasContent, _ := page.Eval(`() => {
		const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
		if (overlays.length > 0) {
			const lastOverlay = overlays[overlays.length - 1];
			const input = lastOverlay.querySelector('.msg-form__contenteditable, div[contenteditable="true"]');
			const text = input ? input.textContent.trim() : '';
			return text.length > 0;
		}
		return false;
	}`)

	if hasContent == nil || !hasContent.Value.Bool() {
		m.logger.Warn("Message not typed via keyboard, trying innerHTML in LAST overlay...")
		// Alternative: Set innerHTML directly in the LAST overlay
		escapedMsg := strings.ReplaceAll(message, "'", "\\'")
		page.Eval(fmt.Sprintf(`() => {
			const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
			if (overlays.length > 0) {
				const lastOverlay = overlays[overlays.length - 1];
				const input = lastOverlay.querySelector('.msg-form__contenteditable, div[contenteditable="true"]');
				if (input) {
					input.focus();
					input.innerHTML = '<p>%s</p>';
					input.dispatchEvent(new Event('input', { bubbles: true }));
					input.dispatchEvent(new Event('change', { bubbles: true }));
				}
			}
		}`, escapedMsg))
		time.Sleep(1 * time.Second)
	} else {
		m.logger.Info("✅ Message typed successfully")
	}

	// Click Send button in the LAST overlay
	m.logger.Debug("Looking for Send button in LAST overlay...")

	sendClicked := false

	// Method 1: Find Send button in the LAST overlay specifically
	result, _ = page.Eval(`() => {
		const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
		if (overlays.length > 0) {
			const lastOverlay = overlays[overlays.length - 1];
			const sendBtn = lastOverlay.querySelector('.msg-form__send-button');
			if (sendBtn && !sendBtn.disabled) {
				sendBtn.click();
				return true;
			}
		}
		return false;
	}`)
	if result != nil && result.Value.Bool() {
		sendClicked = true
		m.logger.Info("✅ Clicked Send button via class selector in LAST overlay")
	}

	// Method 2: Find Send button by text in LAST overlay
	if !sendClicked {
		result, _ := page.Eval(`() => {
			const overlays = document.querySelectorAll('.msg-overlay-conversation-bubble');
			if (overlays.length > 0) {
				const lastOverlay = overlays[overlays.length - 1];
				const buttons = lastOverlay.querySelectorAll('button');
				for (const btn of buttons) {
					if (btn.textContent.trim() === 'Send' && !btn.disabled) {
						btn.click();
						return true;
					}
				}
			}
			return false;
		}`)
		if result != nil && result.Value.Bool() {
			sendClicked = true
			m.logger.Info("✅ Clicked Send button via text match in LAST overlay")
		}
	}

	// Method 3: Direct element click
	if !sendClicked {
		sendBtn, err := page.Timeout(3 * time.Second).Element("button.msg-form__send-button")
		if err == nil && sendBtn != nil {
			if clickErr := sendBtn.Click("left", 1); clickErr == nil {
				sendClicked = true
				m.logger.Info("✅ Clicked Send button via element click")
			}
		}
	}

	if !sendClicked {
		return fmt.Errorf("failed to click Send button - button may be disabled (message not typed?)")
	}

	time.Sleep(2 * time.Second)
	m.logger.Info("✅ Message sent successfully!")

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

	// First, get accepted connections that haven't been messaged yet
	accepted, err := m.db.GetAcceptedConnectionsWithoutMessage()
	if err != nil {
		m.logger.Warnf("Failed to get accepted connections: %v", err)
	}

	// Start with already accepted but not messaged connections
	var newConnections []*storage.ConnectionRequest
	if len(accepted) > 0 {
		m.logger.Infof("Found %d accepted connections awaiting messages", len(accepted))
		newConnections = append(newConnections, accepted...)
	}

	// Get pending connection requests from database
	pending, err := m.db.GetPendingConnectionRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}

	if len(pending) == 0 {
		m.logger.Info("No pending connection requests to check")
		if len(newConnections) > 0 {
			m.logger.Infof("Found %d confirmed new connections", len(newConnections))
			return newConnections, nil
		}
		return nil, nil
	}

	// Check each pending request
	for _, request := range pending {
		// Navigate to profile
		if err := page.Navigate(request.ProfileURL); err != nil {
			m.logger.Warnf("Failed to navigate to profile %s: %v", request.ProfileURL, err)
			continue
		}

		time.Sleep(2 * time.Second)

		// Check connection status using JavaScript - more reliable
		result, err := page.Eval(`() => {
			// Method 1: Check for "1st" degree connection indicator
			var degreeIndicator = document.body.innerText;
			var is1stDegree = degreeIndicator.indexOf('1st') !== -1 && 
			                  (degreeIndicator.indexOf('1st degree connection') !== -1 || 
			                   degreeIndicator.indexOf('· 1st') !== -1);
			
			// Method 2: Check if Message button is in primary actions (only for connections)
			var profileActions = document.querySelector('.pvs-profile-actions, .pv-top-card-v2-ctas');
			var hasMessageButton = false;
			var hasPendingButton = false;
			var hasConnectButton = false;
			
			if (profileActions) {
				var buttons = profileActions.querySelectorAll('button');
				for (var i = 0; i < buttons.length; i++) {
					var text = buttons[i].innerText.trim().toLowerCase();
					var ariaLabel = (buttons[i].getAttribute('aria-label') || '').toLowerCase();
					
					if (text === 'message' || ariaLabel.indexOf('message') !== -1) {
						hasMessageButton = true;
					}
					if (text === 'pending') {
						hasPendingButton = true;
					}
					if (text === 'connect' || ariaLabel.indexOf('connect') !== -1) {
						hasConnectButton = true;
					}
				}
			}
			
			// Method 3: Check for "Pending" button (request sent but not accepted)
			var allButtons = document.querySelectorAll('button');
			for (var j = 0; j < allButtons.length; j++) {
				if (allButtons[j].innerText.trim().toLowerCase() === 'pending') {
					hasPendingButton = true;
					break;
				}
			}
			
			return {
				is1stDegree: is1stDegree,
				hasMessageButton: hasMessageButton,
				hasPendingButton: hasPendingButton,
				hasConnectButton: hasConnectButton,
				isConnected: (is1stDegree || hasMessageButton) && !hasPendingButton
			};
		}`)

		if err != nil {
			m.logger.Warnf("Failed to check connection status for %s: %v", request.ProfileName, err)
			continue
		}

		if result.Value.Val() != nil {
			data := result.Value.Val().(map[string]interface{})
			isConnected, _ := data["isConnected"].(bool)
			hasPending, _ := data["hasPendingButton"].(bool)

			if hasPending {
				m.logger.Debugf("%s: Request still pending", request.ProfileName)
				continue
			}

			if isConnected {
				m.logger.Infof("✅ Confirmed connection: %s", request.ProfileName)
				// Update status in database
				m.db.UpdateConnectionRequestStatus(request.ProfileURL, "accepted")
				newConnections = append(newConnections, request)
			} else {
				m.logger.Debugf("%s: Not connected yet (may have been rejected or expired)", request.ProfileName)
			}
		}
	}

	m.logger.Infof("Found %d confirmed new connections", len(newConnections))
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
