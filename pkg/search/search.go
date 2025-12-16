package search

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/saurabhkuntal/subspace/pkg/stealth"
	"github.com/saurabhkuntal/subspace/pkg/storage"
	"github.com/sirupsen/logrus"
)

// Searcher handles LinkedIn search operations
type Searcher struct {
	stealthMgr *stealth.Manager
	db         *storage.DB
	logger     *logrus.Logger
}

// SearchResult represents a search result
type SearchResult struct {
	ProfileURL string
	Name       string
	Title      string
	Company    string
	Location   string
}

// NewSearcher creates a new searcher
func NewSearcher(stealthMgr *stealth.Manager, db *storage.DB, logger *logrus.Logger) *Searcher {
	return &Searcher{
		stealthMgr: stealthMgr,
		db:         db,
		logger:     logger,
	}
}

// SearchPeople searches for people on LinkedIn
func (s *Searcher) SearchPeople(page *rod.Page, keywords []string, location string, maxResults int) ([]*SearchResult, error) {
	s.logger.Infof("Searching for people: keywords=%v, location=%s", keywords, location)

	// Warm-up phase: visit LinkedIn homepage first to establish session
	s.logger.Info("Warming up session - visiting LinkedIn homepage...")
	if err := page.Navigate("https://www.linkedin.com/feed/"); err == nil {
		time.Sleep(time.Duration(3+rand.Intn(3)) * time.Second)

		// Simulate reading the feed - random scrolls
		for i := 0; i < 2; i++ {
			page.Mouse.Scroll(0, float64(200+rand.Intn(200)), 5)
			time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
		}
		time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)
	}

	s.logger.Info("Session warmed up, proceeding to search...")

	// Build search URL
	searchURL := s.buildSearchURL(keywords, location)
	s.logger.Debugf("Search URL: %s", searchURL)

	// Navigate to search page with retry logic
	maxRetries := 3
	var navErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			s.logger.Infof("Retrying navigation (attempt %d/%d)...", i+1, maxRetries)
			time.Sleep(time.Duration(2+i*2) * time.Second) // Increasing backoff
		}

		navErr = page.Timeout(30 * time.Second).Navigate(searchURL)
		if navErr == nil {
			s.logger.Debug("Navigation successful")
			break
		}

		s.logger.Warnf("Navigation attempt %d failed: %v", i+1, navErr)
	}

	if navErr != nil {
		return nil, fmt.Errorf("failed to navigate to search page after %d attempts: %w", maxRetries, navErr)
	}

	// Log current page URL for debugging
	currentURL := page.MustInfo().URL
	s.logger.Debugf("Current page URL: %s", currentURL)

	// Check if we're on mobile version and redirect to desktop
	if strings.Contains(currentURL, "linkedin.com/m/") {
		s.logger.Warn("Detected mobile version, redirecting to desktop...")
		desktopURL := strings.Replace(currentURL, "linkedin.com/m/", "linkedin.com/", 1)
		if err := page.Navigate(desktopURL); err != nil {
			s.logger.Warnf("Failed to navigate to desktop version: %v", err)
		} else {
			time.Sleep(3 * time.Second)
			currentURL = page.MustInfo().URL
			s.logger.Infof("Redirected to desktop: %s", currentURL)
		}
	}

	s.stealthMgr.HumanDelay()

	// Random scroll to load results
	if err := s.stealthMgr.RandomScroll(page); err != nil {
		s.logger.Warnf("Failed to scroll: %v", err)
	}

	s.stealthMgr.RandomDelay()

	// Collect results from multiple pages
	var allResults []*SearchResult
	pageNum := 1

	for len(allResults) < maxResults {
		s.logger.Infof("Collecting results from page %d...", pageNum)

		// Extract results from current page
		results, err := s.extractResultsFromPage(page)
		if err != nil {
			s.logger.Warnf("Failed to extract results: %v", err)
			break
		}

		if len(results) == 0 {
			s.logger.Info("No more results found")
			break
		}

		// Filter out duplicates
		for _, result := range results {
			if !s.isDuplicate(result, allResults) {
				allResults = append(allResults, result)
				if len(allResults) >= maxResults {
					break
				}
			}
		}

		// Check if we need more results
		if len(allResults) >= maxResults {
			break
		}

		// Try to go to next page
		if !s.goToNextPage(page) {
			s.logger.Info("No more pages available")
			break
		}

		s.stealthMgr.HumanDelay()
		pageNum++
	}

	// Save search history
	searchHistory := &storage.SearchHistory{
		Query:       strings.Join(keywords, ", "),
		Location:    location,
		ResultCount: len(allResults),
		SearchedAt:  time.Now(),
	}
	if err := s.db.SaveSearchHistory(searchHistory); err != nil {
		s.logger.Warnf("Failed to save search history: %v", err)
	}

	s.logger.Infof("Found %d results", len(allResults))
	return allResults, nil
}

// buildSearchURL builds a LinkedIn search URL
func (s *Searcher) buildSearchURL(keywords []string, location string) string {
	// Use desktop version explicitly
	baseURL := "https://www.linkedin.com/search/results/people/"

	params := url.Values{}

	// Add keywords
	if len(keywords) > 0 {
		params.Add("keywords", strings.Join(keywords, " "))
	}

	// Add location
	if location != "" {
		geoUrn := s.getGeoURN(location)
		if geoUrn != "" {
			params.Add("geoUrn", geoUrn)
		}
	}

	// Add filters for connections (2nd and 3rd degree)
	params.Add("network", "[\"S\",\"O\"]")

	// Force desktop origin
	params.Add("origin", "GLOBAL_SEARCH_HEADER")

	return baseURL + "?" + params.Encode()
}

// getGeoURN converts a location string to a LinkedIn geo URN
// This is a simplified version - in production, you'd need a mapping table
func (s *Searcher) getGeoURN(location string) string {
	// Normalize location string for comparison
	locationLower := strings.ToLower(strings.TrimSpace(location))

	// Common locations mapping (case-insensitive keys)
	locationMap := map[string]string{
		"united states": "103644278",
		"usa":           "103644278",
		"san francisco": "90009706",
		"new york":      "105080838",
		"london":        "102257491",
		"india":         "102713980",
		"bangalore":     "104769905",
		"bengaluru":     "104769905",
		"delhi":         "105214831",
		"new delhi":     "105214831",
		"mumbai":        "102713981",
		"hyderabad":     "106449469",
		"chennai":       "102713982",
		"pune":          "106057199",
		"kolkata":       "105183972",
		"singapore":     "102454443",
		"dubai":         "104305776",
		"toronto":       "100025096",
		"sydney":        "104769905",
	}

	if urn, ok := locationMap[locationLower]; ok {
		s.logger.Debugf("Mapped location '%s' to URN: %s", location, urn)
		return urn
	}

	s.logger.Warnf("Location '%s' not in mapping, search will be without location filter", location)
	return ""
}

// extractResultsFromPage extracts search results from the current page
func (s *Searcher) extractResultsFromPage(page *rod.Page) ([]*SearchResult, error) {
	// Initial wait for page to start loading
	s.logger.Debug("Waiting for search results to load...")
	time.Sleep(time.Duration(6+rand.Intn(3)) * time.Second)

	// Simulate human reading behavior with pauses
	s.logger.Info("Simulating human reading behavior...")
	time.Sleep(time.Duration(800+rand.Intn(400)) * time.Millisecond)

	// Random pauses mimicking reading
	for i := 0; i < 4; i++ {
		time.Sleep(time.Duration(600+rand.Intn(600)) * time.Millisecond)
	}

	// Human-like scrolling with natural variations - MORE EXTENSIVE
	s.logger.Info("Extended scrolling to load results (this may take 40-60 seconds)...")
	for i := 0; i < 10; i++ {
		// Vary scroll amount naturally
		scrollAmount := 200 + rand.Intn(250)
		scrollSpeed := 5 + rand.Intn(7)
		page.Mouse.Scroll(0, float64(scrollAmount), scrollSpeed)

		// Variable pause between scrolls (2-5 seconds)
		pause := time.Duration(2000+rand.Intn(3000)) * time.Millisecond
		time.Sleep(pause)

		// More frequently scroll up slightly (mimics reading)
		if i%2 == 0 {
			page.Mouse.Scroll(0, float64(-60-rand.Intn(60)), 4)
			time.Sleep(time.Duration(1500+rand.Intn(1500)) * time.Millisecond)
		}

		// Longer pause at some points
		if i == 4 || i == 7 {
			s.logger.Debug("Pausing to simulate reading...")
			time.Sleep(time.Duration(4000+rand.Intn(3000)) * time.Millisecond)
		}

		// Check periodically if real results have loaded
		if i > 3 && i%3 == 0 {
			realCount, _ := page.Eval(`() => {
				const links = document.querySelectorAll('a[href*="/in/"]');
				let count = 0;
				links.forEach(link => {
					if (!link.closest('[data-chameleon-result-urn*="headless"]')) {
						count++;
					}
				});
				return count;
			}`)
			if realCount.Value.Int() > 0 {
				s.logger.Infof("Detected %d real profile links, continuing...", realCount.Value.Int())
			}
		}
	}

	// Scroll to top and back down (common human behavior)
	s.logger.Debug("Scrolling to top and re-reading...")
	page.Mouse.Scroll(0, -800, 5)
	time.Sleep(time.Duration(2000+rand.Intn(1500)) * time.Millisecond)
	page.Mouse.Scroll(0, 400, 4)
	time.Sleep(time.Duration(1500+rand.Intn(1000)) * time.Millisecond)

	// Final extended wait for all lazy-loaded content
	s.logger.Debug("Waiting for all content to render...")
	time.Sleep(time.Duration(6+rand.Intn(4)) * time.Second)

	// Additional human behavior: pause as if considering filters
	s.logger.Debug("Simulating interest in page elements...")
	time.Sleep(time.Duration(800+rand.Intn(400)) * time.Millisecond)

	// Log current URL and page title for debugging
	currentURL := page.MustInfo().URL
	pageTitle, _ := page.Eval(`() => document.title`)
	s.logger.Infof("Current URL: %s", currentURL)
	s.logger.Infof("Page Title: %v", pageTitle)

	// Take a screenshot for debugging (saved to screenshots directory)
	screenshotPath := fmt.Sprintf("screenshots/search_%d.png", time.Now().Unix())
	page.MustScreenshot(screenshotPath)
	s.logger.Infof("Screenshot saved to: %s", screenshotPath)

	// Check if we hit an error or restriction page
	html, _ := page.HTML()
	if strings.Contains(strings.ToLower(html), "we've detected unusual activity") ||
		strings.Contains(strings.ToLower(html), "please verify") ||
		strings.Contains(strings.ToLower(html), "security verification") {
		s.logger.Error("LinkedIn has detected automation - please complete any verification manually")
		return nil, fmt.Errorf("LinkedIn security check detected")
	}

	// Check if LinkedIn is showing the skeleton loader indicator
	skeletonCheck, _ := page.Eval(`() => {
		const skeletons = document.querySelectorAll('[data-chameleon-result-urn*="headless"]');
		const realProfiles = document.querySelectorAll('a[href*="/in/"]:not([href*="headless"])');
		return {
			skeletonCount: skeletons.length,
			realProfileLinks: realProfiles.length,
			hasLoadingIndicator: document.querySelector('.artdeco-loader') !== null
		};
	}`)
	s.logger.Infof("Page state check: %v", skeletonCheck)

	// Additional wait after detecting page state
	s.logger.Debug("Page state analyzed, waiting a bit more...")
	time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)

	// Try different selectors for search results container
	selectors := []string{
		".search-results-container",
		"ul.reusable-search__entity-result-list",
		"div.search-results-container",
		"main[class*='scaffold-layout']",
		".scaffold-layout__list",
		"[data-view-name='search-entity-result-universal-template']",
		".entity-result",
		"main.scaffold-layout__main",
		"div.search-results",
		"ul[class*='reusable-search']",
	}

	var foundSelector string
	for _, selector := range selectors {
		element, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil && element != nil {
			foundSelector = selector
			s.logger.Debugf("Found results container with selector: %s", selector)
			break
		}
	}

	// Even if container not found, try to extract results directly
	if foundSelector == "" {
		s.logger.Warn("Could not find search results container with any known selector")
		s.logger.Info("Trying direct result extraction anyway...")
	}

	// Find all result items with multiple possible selectors
	var elements rod.Elements
	var err error

	resultSelectors := []string{
		"li.reusable-search__result-container",
		".reusable-search__result-container",
		"li[data-view-name='search-entity-result-universal-template']",
		"[data-view-name='search-entity-result-universal-template']",
		".entity-result",
		"div.entity-result",
		"li.search-result",
		".search-result__wrapper",
		"div[class*='search-result']",
	}

	for _, selector := range resultSelectors {
		elements, err = page.Elements(selector)
		if err == nil && len(elements) > 0 {
			s.logger.Infof("✓ Found %d results with selector: %s", len(elements), selector)
			break
		} else {
			s.logger.Debugf("✗ Selector '%s' found 0 results", selector)
		}
	}

	if len(elements) == 0 {
		s.logger.Warn("No result elements found with standard selectors")
		s.logger.Info("Attempting fallback link extraction method...")
		return s.extractResultsFromLinks(page)
	}

	var results []*SearchResult

	s.logger.Infof("Starting extraction from %d elements...", len(elements))

	// Filter out skeleton/placeholder elements (those with "headless" URN)
	realElements := make([]*rod.Element, 0)
	for _, element := range elements {
		urn, err := element.Attribute("data-chameleon-result-urn")
		if err == nil && urn != nil && !strings.Contains(*urn, "headless") {
			realElements = append(realElements, element)
		} else if err != nil || urn == nil {
			// If no URN attribute, might be a real element - include it
			realElements = append(realElements, element)
		}
	}

	s.logger.Infof("Found %d real profile elements (filtered out %d skeleton placeholders)",
		len(realElements), len(elements)-len(realElements))

	// If no real elements found, try more aggressive extraction
	if len(realElements) == 0 {
		s.logger.Warn("LinkedIn anti-bot detection active - all results are skeleton placeholders")
		s.logger.Warn("💡 TIP: Try running a warm-up session (Option 7) or clear session with --clear-session flag")
		s.logger.Info("Attempting JavaScript-based extraction...")

		// Try waiting longer and scrolling more - sometimes content loads late
		s.logger.Info("Waiting additional time for content to load...")
		time.Sleep(5 * time.Second)

		// Try scrolling to trigger lazy loading
		for i := 0; i < 3; i++ {
			page.Mouse.Scroll(0, float64(300+rand.Intn(200)), 5)
			time.Sleep(time.Duration(2000+rand.Intn(1000)) * time.Millisecond)
		}

		// Check again for real profile links
		realLinkCount, _ := page.Eval(`() => {
			const links = document.querySelectorAll('a[href*="/in/"]');
			let count = 0;
			links.forEach(link => {
				const text = link.innerText.trim();
				if (text && text.length > 2 && text.length < 100 && !text.includes('LinkedIn Member')) {
					count++;
				}
			});
			return count;
		}`)

		if realLinkCount.Value.Int() > 0 {
			s.logger.Infof("Found %d real profile links after additional wait", realLinkCount.Value.Int())
		}

		return s.extractResultsViaJavaScript(page)
	}

	for i, element := range realElements {
		s.logger.Infof("Processing element %d/%d...", i+1, len(realElements))

		result, err := s.extractResult(element)
		if err != nil {
			s.logger.Infof("❌ Failed to extract result %d: %v", i+1, err)
			continue
		}

		if result != nil && result.ProfileURL != "" {
			s.logger.Infof("✓ Extracted %d: Name='%s', Title='%s', URL='%s'",
				i+1, result.Name, result.Title, result.ProfileURL)
			results = append(results, result)
		} else {
			s.logger.Infof("⊘ Skipped result %d: empty profile URL", i+1)
		}
	}

	s.logger.Infof("Successfully extracted %d profiles from %d elements", len(results), len(elements))
	return results, nil
}

// extractResultsViaJavaScript tries to extract results using JavaScript evaluation
func (s *Searcher) extractResultsViaJavaScript(page *rod.Page) ([]*SearchResult, error) {
	s.logger.Debug("Using JavaScript-based extraction method")

	result, err := page.Eval(`() => {
		var results = [];
		var seen = {};
		
		// Find all profile links and their context
		var profileLinks = document.querySelectorAll('a[href*="/in/"]');
		
		for (var i = 0; i < profileLinks.length; i++) {
			var link = profileLinks[i];
			var href = link.href;
			if (!href || seen[href]) continue;
			
			// Clean URL
			var url = href.split('?')[0];
			if (seen[url]) continue;
			seen[url] = true;
			
			// Skip non-profile links
			if (url.indexOf('/company/') !== -1) continue;
			if (url.indexOf('/school/') !== -1) continue;
			if (url.indexOf('/posts/') !== -1) continue;
			
			// Find the closest search result container
			var container = link.closest('li, div.entity-result, [data-view-name]') || link.parentElement;
			if (!container) continue;
			
			// Extract name - try multiple approaches
			var name = '';
			var nameElement = container.querySelector('span[dir="ltr"] span[aria-hidden="true"], .entity-result__title-text a, .actor-name');
			if (nameElement) {
				name = nameElement.innerText.trim();
			} else {
				name = link.innerText.trim();
			}
			
			// If no name, try to extract from URL
			if (!name || name.length < 2) {
				var match = url.match(/\/in\/([^\/]+)/);
				if (match && match[1]) {
					name = match[1].replace(/-/g, ' ').replace(/\d+$/, '').trim();
					name = name.split(' ').map(function(word) {
						return word.charAt(0).toUpperCase() + word.slice(1);
					}).join(' ');
				}
			}
			
			// Skip if name looks invalid
			if (!name || name.length > 100 || name.indexOf('LinkedIn Member') !== -1) continue;
			
			// Extract title
			var title = '';
			var titleElement = container.querySelector('.entity-result__primary-subtitle, .subline-level-1');
			if (titleElement) {
				title = titleElement.innerText.trim();
			}
			
			// Extract location
			var location = '';
			var locationElement = container.querySelector('.entity-result__secondary-subtitle, .subline-level-2');
			if (locationElement) {
				location = locationElement.innerText.trim();
			}
			
			results.push({
				profileURL: url,
				name: name,
				title: title,
				location: location
			});
		}
		
		return results;
	}`)

	if err != nil {
		s.logger.Warnf("JavaScript extraction failed: %v", err)
		return s.extractResultsFromLinks(page)
	}

	var jsResults []map[string]interface{}
	if arr, ok := result.Value.Val().([]interface{}); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				jsResults = append(jsResults, m)
			}
		}
	}

	var results []*SearchResult
	for _, r := range jsResults {
		sr := &SearchResult{
			ProfileURL: getString(r, "profileURL"),
			Name:       getString(r, "name"),
			Title:      getString(r, "title"),
			Location:   getString(r, "location"),
		}
		if sr.ProfileURL != "" && sr.Name != "" {
			results = append(results, sr)
		}
	}

	s.logger.Infof("JavaScript extraction found %d profiles", len(results))

	if len(results) == 0 {
		s.logger.Warn("JavaScript extraction found no results, falling back to link extraction")
		return s.extractResultsFromLinks(page)
	}

	return results, nil
}

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// extractResultsFromLinks is a fallback method that extracts LinkedIn profile links from the page
func (s *Searcher) extractResultsFromLinks(page *rod.Page) ([]*SearchResult, error) {
	s.logger.Info("Using enhanced link extraction method")

	// First, let's debug what's on the page
	debugResult, _ := page.Eval(`() => {
		var links = document.querySelectorAll('a[href*="/in/"]');
		var debug = [];
		for (var i = 0; i < Math.min(links.length, 5); i++) {
			debug.push({
				href: links[i].href,
				text: links[i].innerText.substring(0, 50),
				parentClass: links[i].parentElement ? links[i].parentElement.className : 'none'
			});
		}
		return { total: links.length, samples: debug };
	}`)
	s.logger.Infof("DEBUG - Links found: %v", debugResult)

	// Use Rod's native DOM methods instead of JavaScript evaluation
	links, err := page.Elements(`a[href*="/in/"]`)
	if err != nil {
		s.logger.Warnf("Failed to find profile links: %v", err)
		return nil, err
	}

	s.logger.Infof("Found %d potential profile links via Rod Elements", len(links))

	var results []*SearchResult
	seen := make(map[string]bool)

	for i, link := range links {
		// Get href attribute
		hrefProp, err := link.Property("href")
		if err != nil {
			continue
		}
		href := hrefProp.String()

		// Skip non-profile URLs
		if !strings.Contains(href, "/in/") {
			continue
		}
		if strings.Contains(href, "/company/") || strings.Contains(href, "/school/") || strings.Contains(href, "/posts/") {
			continue
		}

		// Clean URL (remove query params)
		cleanURL := href
		if idx := strings.Index(href, "?"); idx != -1 {
			cleanURL = href[:idx]
		}

		// Skip duplicates
		if seen[cleanURL] {
			continue
		}
		seen[cleanURL] = true

		// Try to get name from link text
		name := ""
		text, err := link.Text()
		if err == nil {
			// The text often contains the full card content with newlines
			// Extract just the first line which is typically the name
			text = strings.TrimSpace(text)
			lines := strings.Split(text, "\n")
			if len(lines) > 0 {
				name = strings.TrimSpace(lines[0])
			}
		}

		// If name is empty or too short, try to find a span with aria-hidden inside the link
		if len(name) < 2 {
			nameSpan, err := link.Element(`span[aria-hidden="true"]`)
			if err == nil && nameSpan != nil {
				spanText, _ := nameSpan.Text()
				name = strings.TrimSpace(spanText)
				// Also take first line if multiline
				if strings.Contains(name, "\n") {
					lines := strings.Split(name, "\n")
					name = strings.TrimSpace(lines[0])
				}
			}
		}

		// If still no name, extract from URL
		if len(name) < 2 {
			// Extract from URL like /in/john-doe-123/
			parts := strings.Split(cleanURL, "/in/")
			if len(parts) > 1 {
				slug := strings.TrimSuffix(parts[1], "/")
				// Remove trailing numbers and convert dashes to spaces
				slug = strings.TrimRight(slug, "0123456789")
				slug = strings.ReplaceAll(slug, "-", " ")
				// Capitalize words
				words := strings.Fields(slug)
				for j, word := range words {
					if len(word) > 0 {
						words[j] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
					}
				}
				name = strings.Join(words, " ")
			}
		}

		// Clean up name - remove common suffixes
		name = strings.TrimSpace(name)
		// Remove connection degree indicators
		if idx := strings.Index(name, " •"); idx > 0 {
			name = strings.TrimSpace(name[:idx])
		}

		// Skip invalid names
		if len(name) < 2 || len(name) > 100 {
			s.logger.Debugf("Skipping link %d: name too short/long (%d chars): %s", i, len(name), name)
			continue
		}
		if strings.Contains(strings.ToLower(name), "linkedin member") {
			continue
		}
		if strings.ToLower(name) == "view profile" {
			continue
		}

		// Try to get title from parent container
		title := ""
		location := ""

		// Find parent li element
		parent, err := link.Parent()
		if err == nil && parent != nil {
			// Go up to find the li container
			for j := 0; j < 5; j++ {
				tagName, _ := parent.Property("tagName")
				if tagName.String() == "LI" {
					// Found the list item, try to get subtitle
					subtitleEl, err := parent.Element(`.entity-result__primary-subtitle`)
					if err == nil && subtitleEl != nil {
						title, _ = subtitleEl.Text()
						title = strings.TrimSpace(title)
					}
					locEl, err := parent.Element(`.entity-result__secondary-subtitle`)
					if err == nil && locEl != nil {
						location, _ = locEl.Text()
						location = strings.TrimSpace(location)
					}
					break
				}
				parent, err = parent.Parent()
				if err != nil {
					break
				}
			}
		}

		result := &SearchResult{
			ProfileURL: cleanURL,
			Name:       name,
			Title:      title,
			Location:   location,
		}

		results = append(results, result)
		s.logger.Infof("✓ Extracted %d: %s (%s)", i+1, name, cleanURL)

		// Limit to 50 results
		if len(results) >= 50 {
			break
		}
	}

	s.logger.Infof("Enhanced link extraction found %d profiles", len(results))
	return results, nil
}

// extractResult extracts a single result from an element
func (s *Searcher) extractResult(element *rod.Element) (result *SearchResult, err error) {
	// Recover from any panics during extraction
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during extraction: %v", r)
		}
	}()

	result = &SearchResult{}

	// Extract profile URL - try multiple selectors
	// IMPORTANT: Order matters! More specific selectors first to avoid catching wrong elements
	var linkElement *rod.Element

	linkSelectors := []string{
		"a[href*='/in/']",                           // Most specific - links containing /in/
		".entity-result__title-text a",              // Profile name link in entity result
		"span.entity-result__title-text a",          // Variant with span wrapper
		"div.entity-result__title-text a",           // Variant with div wrapper
		"a.app-aware-link[href*='/in/']",            // App-aware link that's a profile
		"a[data-test-app-aware-link][href*='/in/']", // Data attribute variant
	}

	for _, selector := range linkSelectors {
		linkElement, err = element.Element(selector)
		if err == nil && linkElement != nil {
			// Verify it's actually a profile link before accepting
			href, err := linkElement.Property("href")
			if err == nil {
				url := href.String()
				if strings.Contains(url, "/in/") && !strings.Contains(url, "/headless") {
					s.logger.Infof("✓ Found valid profile link with selector: %s", selector)
					result.ProfileURL = url
					break
				}
			}
		}
	}

	if result.ProfileURL == "" {
		// Try to get the HTML of this element to see what's inside
		html, _ := element.HTML()
		s.logger.Infof("Element HTML (first 1000 chars): %s", truncate(html, 1000))

		// Try to get all links and log them
		allLinks, _ := element.Elements("a")
		s.logger.Infof("Found %d links in element", len(allLinks))
		for i, link := range allLinks {
			href, err := link.Property("href")
			if err == nil {
				s.logger.Infof("  Link %d: %s", i+1, href.String())
			}
		}

		return nil, fmt.Errorf("could not find valid profile link with any selector")
	}

	// Log the found URL for debugging
	s.logger.Infof("🔍 Found URL: %s", result.ProfileURL)

	// Clean up profile URL (remove query parameters)
	if idx := strings.Index(result.ProfileURL, "?"); idx != -1 {
		result.ProfileURL = result.ProfileURL[:idx]
	}

	// Extract name - try multiple selectors
	nameSelectors := []string{
		"span[dir='ltr'] span[aria-hidden='true']",
		".entity-result__title-text a span[aria-hidden='true']",
		".entity-result__title-text span[aria-hidden='true']",
		".entity-result__title-text a span",
		".entity-result__title-text",
		"span.entity-result__title-line",
	}

	for _, selector := range nameSelectors {
		nameElement, err := element.Element(selector)
		if err == nil {
			name, _ := nameElement.Text()
			name = strings.TrimSpace(name)
			// Filter out "LinkedIn Member" placeholder names
			if name != "" && !strings.Contains(name, "LinkedIn Member") {
				result.Name = name
				break
			}
		}
	}

	// If still no name, try getting it from the link
	if result.Name == "" {
		name, _ := linkElement.Text()
		name = strings.TrimSpace(name)
		if !strings.Contains(name, "LinkedIn Member") {
			result.Name = name
		}
	}

	// Name is required - if we don't have it, this isn't a valid result
	if result.Name == "" {
		return nil, fmt.Errorf("no name found")
	}

	// Use JavaScript to extract all info at once - more reliable than Go selectors
	jsExtract, err := element.Eval(`() => {
		const result = {};
		
		// Get all text content for debugging
		const allText = this.innerText;
		result.allText = allText;
		
		// LinkedIn's search result structure:
		// - Primary subtitle: Job headline/title
		// - Secondary subtitle: Location  
		// - Badge text: Connection degree (• 3rd+)
		
		// Extract connection degree first (it often appears with •)
		const badgeText = this.querySelector('.entity-result__badge-text');
		if (badgeText) {
			result.degree = badgeText.innerText.trim();
		}
		
		// Get the primary subtitle - this SHOULD be the job title
		// But LinkedIn sometimes puts degree info there too
		const primarySubtitle = this.querySelector('.entity-result__primary-subtitle');
		let primaryText = '';
		if (primarySubtitle) {
			primaryText = primarySubtitle.innerText.trim();
		}
		
		// Get secondary subtitle - this SHOULD be the location
		const secondarySubtitle = this.querySelector('.entity-result__secondary-subtitle');
		let secondaryText = '';
		if (secondarySubtitle) {
			secondaryText = secondarySubtitle.innerText.trim();
		}
		
		// Determine what's title vs location based on content
		// Connection degrees start with • and contain "1st", "2nd", "3rd"
		const isDegree = (text) => {
			return text && (text.startsWith('•') || text.includes('1st') || text.includes('2nd') || text.includes('3rd'));
		};
		
		// Check if primary contains degree info or actual title
		if (primaryText && !isDegree(primaryText)) {
			result.title = primaryText;
			result.location = secondaryText;
		} else if (secondaryText && !isDegree(secondaryText)) {
			// Secondary has the title, check for location elsewhere
			result.title = secondaryText;
			result.location = '';
		}
		
		// If we still don't have a title, parse from full text
		if (!result.title) {
			const lines = allText.split('\\n').map(l => l.trim()).filter(l => l && l.length > 3);
			// Skip name (first line) and degree info, find title-like content
			for (let i = 1; i < Math.min(lines.length, 6); i++) {
				const line = lines[i];
				if (isDegree(line) || line.includes('View') || line.includes('Message') || 
				    line.includes('Connect') || line.length < 5) {
					continue;
				}
				// Good title candidate: contains job-related words or is reasonably long
				if (line.length >= 10 && line.length <= 200) {
					result.title = line;
					break;
				}
			}
		}
		
		// Try to get summary/snippet
		const summary = this.querySelector('.entity-result__summary');
		if (summary) {
			result.summary = summary.innerText.trim();
		}
		
		return result;
	}`)

	if err == nil && jsExtract.Value.Val() != nil {
		if jsData, ok := jsExtract.Value.Val().(map[string]interface{}); ok {
			if title, ok := jsData["title"].(string); ok && title != "" && result.Title == "" {
				result.Title = title
			}
			if location, ok := jsData["location"].(string); ok && location != "" && result.Location == "" {
				result.Location = location
			}
			// Log the all text for debugging if we still don't have title
			if result.Title == "" {
				if allText, ok := jsData["allText"].(string); ok {
					s.logger.Debugf("Element text for %s: %s", result.Name, truncate(allText, 300))
				}
			}
		}
	}

	return result, nil
}

// goToNextPage attempts to navigate to the next page of results
func (s *Searcher) goToNextPage(page *rod.Page) bool {
	// Look for "Next" button
	nextButton, err := page.Timeout(5 * time.Second).Element("button[aria-label='Next']")
	if err != nil {
		return false
	}

	// Check if button is disabled
	disabled, err := nextButton.Property("disabled")
	if err == nil && disabled.Bool() {
		return false
	}

	// Click next button
	if err := s.stealthMgr.HumanClick(page, nextButton); err != nil {
		s.logger.Warnf("Failed to click next button: %v", err)
		return false
	}

	// Wait for new results to load
	time.Sleep(3 * time.Second)

	return true
}

// isDuplicate checks if a result is a duplicate
func (s *Searcher) isDuplicate(result *SearchResult, existing []*SearchResult) bool {
	for _, r := range existing {
		if r.ProfileURL == result.ProfileURL {
			return true
		}
	}
	return false
}

// SearchByJobTitle searches for people by job title
func (s *Searcher) SearchByJobTitle(page *rod.Page, jobTitle, location string, maxResults int) ([]*SearchResult, error) {
	return s.SearchPeople(page, []string{jobTitle}, location, maxResults)
}

// SearchByCompany searches for people by company
func (s *Searcher) SearchByCompany(page *rod.Page, company, location string, maxResults int) ([]*SearchResult, error) {
	// Build search URL with company filter
	baseURL := "https://www.linkedin.com/search/results/people/"
	params := url.Values{}
	params.Add("keywords", company)
	params.Add("network", "[\"S\",\"O\"]")

	searchURL := baseURL + "?" + params.Encode()

	if err := page.Navigate(searchURL); err != nil {
		return nil, fmt.Errorf("failed to navigate to search page: %w", err)
	}

	s.stealthMgr.HumanDelay()

	// Extract results
	return s.extractResultsFromPage(page)
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
