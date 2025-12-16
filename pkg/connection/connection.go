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

// clickSendWithoutNote clicks the "Send without a note" button or handles direct send
func (m *Manager) clickSendWithoutNote(page *rod.Page) error {
	m.logger.Debug("Waiting for connection modal to appear...")

	// Wait for modal to appear - try multiple times
	var modalFound bool
	for attempt := 0; attempt < 5; attempt++ {
		time.Sleep(1 * time.Second)

		// Check if modal is present
		checkResult, _ := page.Eval(`() => {
			var modal = document.querySelector('.artdeco-modal, [role="dialog"], [data-test-modal]');
			var hasInvitationText = document.body.innerText.indexOf('Add a note to your invitation') !== -1;
			var hasSendButton = document.body.innerText.indexOf('Send without a note') !== -1;
			return {
				modalFound: !!modal,
				hasInvitationText: hasInvitationText,
				hasSendButton: hasSendButton
			};
		}`)

		if checkResult != nil && checkResult.Value.Val() != nil {
			data := checkResult.Value.Val().(map[string]interface{})
			if hasSend, ok := data["hasSendButton"].(bool); ok && hasSend {
				m.logger.Debug("Modal with Send button detected!")
				modalFound = true
				break
			}
			if hasModal, ok := data["modalFound"].(bool); ok && hasModal {
				m.logger.Debug("Modal detected, checking for buttons...")
				modalFound = true
				break
			}
		}
		m.logger.Debugf("Waiting for modal... attempt %d/5", attempt+1)
	}

	// Take screenshot for debugging
	page.MustScreenshot("screenshots/connection_modal.png")
	m.logger.Debug("Screenshot saved: screenshots/connection_modal.png")

	if !modalFound {
		m.logger.Debug("No modal found after waiting")
	}

	// METHOD 1: Use Rod to find "Send without a note" button
	m.logger.Debug("Looking for Send without a note button using Rod...")
	sendBtn, err := page.Timeout(3*time.Second).ElementR("button", "Send without a note")
	if err == nil && sendBtn != nil {
		visible, _ := sendBtn.Visible()
		if visible {
			m.logger.Debug("Found Send without a note button via Rod, clicking...")
			err = sendBtn.Click(proto.InputMouseButtonLeft, 1)
			if err == nil {
				m.logger.Info("✅ Clicked Send without a note button via Rod")
				return nil
			}
		}
	}

	// METHOD 2: Find buttons and look for exact text match
	m.logger.Debug("Trying button element search...")
	buttons, err := page.Elements("button")
	if err == nil {
		for _, btn := range buttons {
			text, _ := btn.Text()
			text = strings.TrimSpace(text)
			if text == "Send without a note" {
				visible, _ := btn.Visible()
				if visible {
					m.logger.Debug("Found button with exact text, clicking...")
					err = btn.Click(proto.InputMouseButtonLeft, 1)
					if err == nil {
						m.logger.Info("✅ Clicked Send without a note via element search")
						return nil
					}
				}
			}
		}
	}

	// METHOD 3: Try JavaScript click
	m.logger.Debug("Trying JavaScript for Send without a note...")
	result, err := page.Eval(`() => {
		var allButtons = document.querySelectorAll('button');
		var buttonTexts = [];
		for (var i = 0; i < allButtons.length; i++) {
			var txt = allButtons[i].innerText.trim();
			if (txt.length > 0 && txt.length < 50) {
				buttonTexts.push(txt);
			}
		}
		
		// Look for "Send without a note" button
		for (var i = 0; i < allButtons.length; i++) {
			var btn = allButtons[i];
			var text = btn.innerText.trim();
			if (text === 'Send without a note') {
				var rect = btn.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0) {
					btn.click();
					return {success: true, method: 'exact_match', text: text};
				}
			}
		}
		
		// Try partial match
		for (var i = 0; i < allButtons.length; i++) {
			var btn = allButtons[i];
			var text = btn.innerText.trim().toLowerCase();
			if (text.indexOf('send without') !== -1 || text.indexOf('without a note') !== -1) {
				var rect = btn.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0) {
					btn.click();
					return {success: true, method: 'partial_match', text: text};
				}
			}
		}
		
		// Look for primary button with "send" in modal
		var modals = document.querySelectorAll('.artdeco-modal, [role="dialog"]');
		for (var j = 0; j < modals.length; j++) {
			var modal = modals[j];
			var modalBtns = modal.querySelectorAll('button');
			for (var k = 0; k < modalBtns.length; k++) {
				var mbtn = modalBtns[k];
				var mtext = mbtn.innerText.trim().toLowerCase();
				if (mtext.indexOf('send') !== -1 && mtext.indexOf('add') === -1) {
					mbtn.click();
					return {success: true, method: 'modal_send', text: mtext};
				}
			}
		}
		
		return {success: false, allButtons: buttonTexts, modalCount: modals.length};
	}`)

	if err == nil && result.Value.Val() != nil {
		data := result.Value.Val().(map[string]interface{})
		if success, ok := data["success"].(bool); ok && success {
			m.logger.Infof("✅ Clicked Send button via JavaScript (method: %v)", data["method"])
			return nil
		}
		if buttons, ok := data["allButtons"].([]interface{}); ok && len(buttons) > 0 {
			m.logger.Warnf("Available buttons on page: %v", buttons)
		}
	}

	// METHOD 4: Check if connection was already sent (Pending button)
	alreadySentCheck, _ := page.Eval(`() => {
		var actionButtons = document.querySelectorAll('.pvs-profile-actions button, .pv-top-card-v2-ctas button, button');
		for (var i = 0; i < actionButtons.length; i++) {
			var text = actionButtons[i].innerText.trim().toLowerCase();
			if (text === 'pending') {
				return {alreadySent: true, reason: 'pending_button'};
			}
		}
		
		var toasts = document.querySelectorAll('.artdeco-toast-item, [role="alert"]');
		for (var j = 0; j < toasts.length; j++) {
			var toastText = toasts[j].innerText.toLowerCase();
			if (toastText.indexOf('invitation sent') !== -1 || toastText.indexOf('request sent') !== -1) {
				return {alreadySent: true, reason: 'toast_message'};
			}
		}
		
		return {alreadySent: false};
	}`)

	if alreadySentCheck != nil && alreadySentCheck.Value.Val() != nil {
		data := alreadySentCheck.Value.Val().(map[string]interface{})
		if sent, ok := data["alreadySent"].(bool); ok && sent {
			reason := data["reason"]
			m.logger.Infof("Connection request was sent (detected via: %v)", reason)
			return nil
		}
	}

	return fmt.Errorf("send without a note button not found")
}

// findConnectButton finds the Connect button on a profile
// LinkedIn has multiple layouts:
// 1. Direct "Connect" button visible on profile (main action area)
// 2. "Connect" hidden under "More" dropdown menu
// 3. "Follow" is primary, need to use More -> Connect
func (m *Manager) findConnectButton(page *rod.Page) (*rod.Element, error) {
	// Wait for page to load
	time.Sleep(2 * time.Second)

	// Take a screenshot for debugging
	page.MustScreenshot("screenshots/profile_page.png")
	m.logger.Debug("Screenshot saved: screenshots/profile_page.png")

	// FIRST: Try the More dropdown approach - this is where Connect usually is
	m.logger.Debug("Checking More dropdown for Connect option...")

	moreResult, err := page.Eval(`() => {
		// Find the More button in the profile actions area (not sidebar)
		var profileActions = document.querySelector('.pvs-profile-actions, .pv-top-card-v2-ctas, [class*="profile-actions"]');
		var moreBtn = null;
		
		if (profileActions) {
			var btns = profileActions.querySelectorAll('button');
			for (var i = 0; i < btns.length; i++) {
				var text = btns[i].innerText.trim().toLowerCase();
				var ariaLabel = (btns[i].getAttribute('aria-label') || '').toLowerCase();
				if (text === 'more' || ariaLabel.indexOf('more action') !== -1) {
					moreBtn = btns[i];
					break;
				}
			}
		}
		
		// If not found in profile actions, try all buttons
		if (!moreBtn) {
			var allBtns = document.querySelectorAll('button');
			for (var j = 0; j < allBtns.length; j++) {
				var txt = allBtns[j].innerText.trim().toLowerCase();
				var aria = (allBtns[j].getAttribute('aria-label') || '').toLowerCase();
				// Look for More button that's near Message/Follow buttons (profile area)
				if (txt === 'more' && aria.indexOf('more action') !== -1) {
					moreBtn = allBtns[j];
					break;
				}
			}
		}
		
		if (moreBtn) {
			moreBtn.click();
			return {clicked: true, method: 'more_button'};
		}
		
		return {clicked: false};
	}`)

	if err == nil && moreResult.Value.Val() != nil {
		data := moreResult.Value.Val().(map[string]interface{})
		if clicked, ok := data["clicked"].(bool); ok && clicked {
			m.logger.Debug("Clicked More button, waiting for dropdown...")
			time.Sleep(2 * time.Second) // Wait for dropdown animation

			// Take screenshot of dropdown for debugging
			page.MustScreenshot("screenshots/dropdown_open.png")
			m.logger.Debug("Screenshot saved: screenshots/dropdown_open.png")

			// METHOD 1: Use Rod to find Connect text in dropdown using ElementR (regex)
			m.logger.Debug("Looking for Connect option in dropdown using Rod...")

			// Try to find element containing exactly "Connect" text
			connectElement, err := page.Timeout(3*time.Second).ElementR("div, span, li", "^Connect$")
			if err == nil && connectElement != nil {
				visible, _ := connectElement.Visible()
				if visible {
					m.logger.Debug("Found Connect element via Rod ElementR, clicking...")
					err = connectElement.Click(proto.InputMouseButtonLeft, 1)
					if err == nil {
						m.logger.Info("✅ Clicked Connect from dropdown via Rod ElementR")
						return nil, nil
					}
					m.logger.Warnf("Click failed: %v, trying parent...", err)
					// Try clicking parent
					parent, _ := connectElement.Parent()
					if parent != nil {
						parent.Click(proto.InputMouseButtonLeft, 1)
						m.logger.Info("✅ Clicked Connect parent element")
						return nil, nil
					}
				}
			}

			// METHOD 2: Find all elements and look for Connect text
			m.logger.Debug("Trying to find Connect via Elements...")
			elements, err := page.Elements("div, span, li, button")
			if err == nil {
				for _, el := range elements {
					text, _ := el.Text()
					text = strings.TrimSpace(text)
					if text == "Connect" {
						visible, _ := el.Visible()
						box, _ := el.Shape()
						// Check if it's in the dropdown area (right side of page, upper area)
						if visible && box != nil && len(box.Quads) > 0 {
							quad := box.Quads[0]
							x := quad[0]
							y := quad[1]
							// Dropdown is typically on right side (x > 800) and upper area (y < 400)
							if x > 700 && y < 400 && y > 100 {
								m.logger.Debugf("Found Connect at x=%.0f, y=%.0f, clicking...", x, y)
								err = el.Click(proto.InputMouseButtonLeft, 1)
								if err == nil {
									m.logger.Info("✅ Clicked Connect from dropdown via element search")
									return nil, nil
								}
							}
						}
					}
				}
			}

			// METHOD 3: Use JavaScript with simulated mouse click
			m.logger.Debug("Trying JavaScript mouse event simulation...")
			jsResult, err := page.Eval(`() => {
				// Find all elements with "Connect" text
				var elements = document.querySelectorAll('*');
				for (var i = 0; i < elements.length; i++) {
					var el = elements[i];
					// Check direct text content (not nested)
					var directText = '';
					for (var j = 0; j < el.childNodes.length; j++) {
						if (el.childNodes[j].nodeType === Node.TEXT_NODE) {
							directText += el.childNodes[j].textContent;
						}
					}
					directText = directText.trim();
					
					if (directText === 'Connect') {
						var rect = el.getBoundingClientRect();
						// Must be visible and in dropdown area
						if (rect.width > 0 && rect.height > 0 && rect.top > 100 && rect.top < 400 && rect.left > 700) {
							// Simulate full click sequence
							var clickEvent = new MouseEvent('click', {
								bubbles: true,
								cancelable: true,
								view: window,
								clientX: rect.left + rect.width/2,
								clientY: rect.top + rect.height/2
							});
							el.dispatchEvent(clickEvent);
							
							// Also try clicking parent
							if (el.parentElement) {
								el.parentElement.click();
							}
							
							return {success: true, method: 'mouse_event', x: rect.left, y: rect.top};
						}
					}
				}
				return {success: false};
			}`)

			if err == nil && jsResult.Value.Val() != nil {
				data := jsResult.Value.Val().(map[string]interface{})
				if success, ok := data["success"].(bool); ok && success {
					m.logger.Infof("✅ Clicked Connect via mouse event simulation at x=%.0f, y=%.0f", data["x"], data["y"])
					return nil, nil
				}
			}

			m.logger.Warn("Could not find/click Connect in dropdown")
		}
	}

	// SECOND: Try direct Connect button in profile actions area only
	m.logger.Debug("Trying direct Connect button in profile actions...")
	result, err := page.Eval(`() => {
		// Only look in the main profile actions area, not sidebar
		var profileActions = document.querySelector('.pvs-profile-actions, .pv-top-card-v2-ctas');
		if (!profileActions) {
			return {success: false, reason: 'no_profile_actions'};
		}
		
		var buttons = profileActions.querySelectorAll('button');
		for (var i = 0; i < buttons.length; i++) {
			var btn = buttons[i];
			var text = btn.innerText.trim().toLowerCase();
			var ariaLabel = (btn.getAttribute('aria-label') || '').toLowerCase();
			
			// Check for Connect button
			if (text === 'connect' || ariaLabel.indexOf('invite') !== -1) {
				var rect = btn.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0) {
					btn.click();
					return {success: true, method: 'profile_action_connect', text: text};
				}
			}
		}
		
		// List available buttons in profile actions for debugging
		var available = [];
		for (var j = 0; j < buttons.length; j++) {
			available.push(buttons[j].innerText.trim().substring(0, 20));
		}
		
		return {success: false, profileButtons: available};
	}`)

	if err == nil && result.Value.Val() != nil {
		data := result.Value.Val().(map[string]interface{})
		if success, ok := data["success"].(bool); ok && success {
			m.logger.Infof("Clicked Connect button (method: %v)", data["method"])
			return nil, nil
		}
		if buttons, ok := data["profileButtons"].([]interface{}); ok {
			m.logger.Debugf("Profile action buttons: %v", buttons)
		}
	}

	return nil, fmt.Errorf("connect button not found - may need to use More dropdown")
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
