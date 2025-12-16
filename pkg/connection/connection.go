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
	time.Sleep(2 * time.Second)

	// Take screenshot for debugging
	page.MustScreenshot("screenshots/connection_modal.png")
	m.logger.Debug("Screenshot saved: screenshots/connection_modal.png")

	// Use JavaScript to find and click the send button - most reliable approach
	result, err := page.Eval(`() => {
		// Wait a bit for modal to fully render
		
		// Method 1: Look for "Send without a note" or "Send" button
		var buttons = document.querySelectorAll('button');
		for (var i = 0; i < buttons.length; i++) {
			var btn = buttons[i];
			var text = btn.innerText.trim().toLowerCase();
			var ariaLabel = (btn.getAttribute('aria-label') || '').toLowerCase();
			
			// Check for send-related buttons
			if (text.indexOf('send without') !== -1 || text === 'send' || text === 'send now' ||
			    ariaLabel.indexOf('send without') !== -1 || ariaLabel.indexOf('send now') !== -1 || 
			    ariaLabel.indexOf('send invitation') !== -1) {
				// Make sure button is visible
				var rect = btn.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0) {
					btn.click();
					return {success: true, method: 'direct_button', text: text};
				}
			}
		}
		
		// Method 2: Look in modal/dialog specifically
		var modals = document.querySelectorAll('.artdeco-modal, [role="dialog"], .send-invite, [data-test-modal]');
		for (var j = 0; j < modals.length; j++) {
			var modal = modals[j];
			var modalBtns = modal.querySelectorAll('button');
			for (var k = 0; k < modalBtns.length; k++) {
				var mbtn = modalBtns[k];
				var mtext = mbtn.innerText.trim().toLowerCase();
				
				if (mtext.indexOf('send') !== -1) {
					mbtn.click();
					return {success: true, method: 'modal_button', text: mtext};
				}
			}
			
			// Try primary button in modal
			var primaryBtn = modal.querySelector('button.artdeco-button--primary, button[class*="primary"]');
			if (primaryBtn) {
				var ptext = primaryBtn.innerText.trim().toLowerCase();
				if (ptext.indexOf('send') !== -1 || ptext.indexOf('connect') !== -1) {
					primaryBtn.click();
					return {success: true, method: 'primary_button', text: ptext};
				}
			}
		}
		
		// Method 3: Find any visible primary/action button with Send text
		var allButtons = document.querySelectorAll('button.artdeco-button--primary, button.artdeco-button--2, button[class*="primary"]');
		for (var m = 0; m < allButtons.length; m++) {
			var abtn = allButtons[m];
			var atext = abtn.innerText.trim().toLowerCase();
			var rect = abtn.getBoundingClientRect();
			
			if (rect.width > 0 && rect.height > 0 && (atext.indexOf('send') !== -1)) {
				abtn.click();
				return {success: true, method: 'primary_class', text: atext};
			}
		}
		
		// Return debug info
		return {
			success: false, 
			availableButtons: Array.from(buttons).slice(0, 20).map(function(b) { 
				return {text: b.innerText.trim().substring(0, 40), classes: b.className.substring(0, 50)}; 
			}),
			modalsFound: modals.length
		};
	}`)

	if err == nil && result.Value.Val() != nil {
		data := result.Value.Val().(map[string]interface{})
		if success, ok := data["success"].(bool); ok && success {
			m.logger.Infof("Clicked Send button via JavaScript (method: %v, text: %v)", data["method"], data["text"])
			return nil
		}
		// Log available buttons for debugging
		if buttons, ok := data["availableButtons"].([]interface{}); ok {
			m.logger.Debugf("Available buttons in modal: %v", buttons)
		}
		if modals, ok := data["modalsFound"].(float64); ok {
			m.logger.Debugf("Modals found: %v", modals)
		}
	}

	// If JS failed, try Rod selectors as fallback
	sendSelectors := []string{
		"button[aria-label='Send without a note']",
		"button[aria-label='Send now']",
		"button[aria-label='Send invitation']",
	}

	for _, selector := range sendSelectors {
		btn, err := page.Timeout(2 * time.Second).Element(selector)
		if err == nil && btn != nil {
			visible, _ := btn.Visible()
			if visible {
				if err := btn.Click(proto.InputMouseButtonLeft, 1); err == nil {
					m.logger.Debug("Clicked Send button via Rod selector")
					return nil
				}
			}
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

	// Take a screenshot for debugging
	page.MustScreenshot("screenshots/profile_page.png")
	m.logger.Debug("Screenshot saved: screenshots/profile_page.png")

	// First try using JavaScript to find and click the Connect button - most reliable
	m.logger.Debug("Trying JavaScript to find Connect button...")
	result, err := page.Eval(`() => {
		// Method 1: Find button with "Connect" text and connect icon
		var buttons = document.querySelectorAll('button');
		for (var i = 0; i < buttons.length; i++) {
			var btn = buttons[i];
			var text = btn.innerText.trim().toLowerCase();
			var ariaLabel = (btn.getAttribute('aria-label') || '').toLowerCase();
			
			// Check for Connect button
			if (text === 'connect' || text.indexOf('connect') === 0 || 
			    ariaLabel.indexOf('connect') !== -1 || ariaLabel.indexOf('invite') !== -1) {
				// Make sure button is visible
				var rect = btn.getBoundingClientRect();
				if (rect.width > 0 && rect.height > 0) {
					console.log('Found Connect button:', text, ariaLabel);
					btn.click();
					return {success: true, method: 'direct_button', text: text};
				}
			}
		}
		
		// Method 2: Find the profile actions section and look for Connect
		var actionSections = document.querySelectorAll('.pvs-profile-actions, .pv-top-card-v2-ctas, [class*="profile-actions"]');
		for (var j = 0; j < actionSections.length; j++) {
			var section = actionSections[j];
			var btns = section.querySelectorAll('button');
			for (var k = 0; k < btns.length; k++) {
				var b = btns[k];
				var txt = b.innerText.trim().toLowerCase();
				if (txt === 'connect' || txt.indexOf('connect') === 0) {
					b.click();
					return {success: true, method: 'action_section', text: txt};
				}
			}
		}
		
		// Method 3: Look for the specific LinkedIn Connect button structure
		// LinkedIn uses a button with a span containing "Connect" and an icon
		var allSpans = document.querySelectorAll('button span');
		for (var m = 0; m < allSpans.length; m++) {
			var span = allSpans[m];
			if (span.innerText.trim().toLowerCase() === 'connect') {
				var parentBtn = span.closest('button');
				if (parentBtn) {
					parentBtn.click();
					return {success: true, method: 'span_parent', text: 'connect'};
				}
			}
		}
		
		return {success: false, availableButtons: Array.from(document.querySelectorAll('button')).slice(0, 15).map(function(b) { 
			return {text: b.innerText.trim().substring(0, 30), aria: b.getAttribute('aria-label')}; 
		})};
	}`)

	if err == nil && result.Value.Val() != nil {
		data := result.Value.Val().(map[string]interface{})
		if success, ok := data["success"].(bool); ok && success {
			m.logger.Infof("Clicked Connect button via JavaScript (method: %v)", data["method"])
			return nil, nil // Return nil to indicate button was already clicked
		}
		// Log available buttons for debugging
		if buttons, ok := data["availableButtons"].([]interface{}); ok {
			m.logger.Debugf("Available buttons on page: %v", buttons)
		}
	}

	m.logger.Debug("Direct Connect not found, checking More dropdown...")

	// Try to find and click the "More" button to reveal Connect option
	moreResult, err := page.Eval(`() => {
		// Find More button
		var buttons = document.querySelectorAll('button');
		for (var i = 0; i < buttons.length; i++) {
			var btn = buttons[i];
			var text = btn.innerText.trim().toLowerCase();
			var ariaLabel = (btn.getAttribute('aria-label') || '').toLowerCase();
			
			if (text === 'more' || ariaLabel.indexOf('more') !== -1) {
				btn.click();
				return {clicked: true, text: text};
			}
		}
		return {clicked: false};
	}`)

	if err == nil && moreResult.Value.Val() != nil {
		data := moreResult.Value.Val().(map[string]interface{})
		if clicked, ok := data["clicked"].(bool); ok && clicked {
			m.logger.Debug("Clicked More button, waiting for dropdown...")
			time.Sleep(1500 * time.Millisecond)

			// Now look for Connect in the dropdown
			dropdownResult, err := page.Eval(`() => {
				// Look for Connect in dropdown menu
				var menuItems = document.querySelectorAll('.artdeco-dropdown__item, [role="menuitem"], .artdeco-dropdown__content-inner li, div[data-control-name]');
				for (var i = 0; i < menuItems.length; i++) {
					var item = menuItems[i];
					var text = item.innerText.trim().toLowerCase();
					
					if (text.indexOf('connect') !== -1 || text === 'add') {
						item.click();
						return {success: true, text: text};
					}
				}
				
				// Also try clicking any visible Connect text
				var allDivs = document.querySelectorAll('div, span, li');
				for (var j = 0; j < allDivs.length; j++) {
					var el = allDivs[j];
					var txt = el.innerText.trim().toLowerCase();
					if (txt === 'connect' && el.offsetParent !== null) {
						el.click();
						return {success: true, text: 'connect from div/span'};
					}
				}
				
				return {success: false, menuItems: Array.from(menuItems).map(function(m) { return m.innerText.trim().substring(0, 30); })};
			}`)

			if err == nil && dropdownResult.Value.Val() != nil {
				data := dropdownResult.Value.Val().(map[string]interface{})
				if success, ok := data["success"].(bool); ok && success {
					m.logger.Infof("Clicked Connect from dropdown: %v", data["text"])
					return nil, nil
				}
				if items, ok := data["menuItems"].([]interface{}); ok {
					m.logger.Debugf("Dropdown menu items: %v", items)
				}
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
