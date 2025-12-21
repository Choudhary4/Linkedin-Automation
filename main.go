package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/saurabhkuntal/subspace/pkg/auth"
	"github.com/saurabhkuntal/subspace/pkg/config"
	"github.com/saurabhkuntal/subspace/pkg/connection"
	"github.com/saurabhkuntal/subspace/pkg/messaging"
	"github.com/saurabhkuntal/subspace/pkg/search"
	"github.com/saurabhkuntal/subspace/pkg/stealth"
	"github.com/saurabhkuntal/subspace/pkg/storage"
	"github.com/sirupsen/logrus"
)

var reader = bufio.NewReader(os.Stdin)

func readInput() string {
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

var (
	configPath   = flag.String("config", "", "Path to configuration YAML file")
	action       = flag.String("action", "", "Action to perform: search, connect, message, detect-connections (leave empty for interactive menu)")
	keywords     = flag.String("keywords", "", "Search keywords (comma-separated)")
	location     = flag.String("location", "", "Search location")
	maxResults   = flag.Int("max-results", 10, "Maximum search results to retrieve")
	message      = flag.String("message", "", "Custom message for connection requests or messages")
	clearSession = flag.Bool("clear-session", false, "Clear saved session before starting")
	interactive  = flag.Bool("interactive", true, "Run in interactive mode (default: true)")
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please ensure .env file exists or provide credentials via environment variables.\n")
		os.Exit(1)
	}

	// Setup logger
	logger := setupLogger(cfg.App.LogLevel)
	logger.Info("Starting SubSpace LinkedIn Automation")
	logger.Warn("⚠️  This tool is for educational purposes only. Using it may violate LinkedIn's Terms of Service.")

	// Initialize database
	db, err := storage.NewDB(cfg.Database.Path)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize stealth manager
	stealthMgr := stealth.NewManager(
		cfg.Stealth.MinActionDelay,
		cfg.Stealth.MaxActionDelay,
		cfg.Stealth.EnableRandomScroll,
		cfg.Stealth.EnableTypingSimulation,
		cfg.Stealth.EnableMouseHover,
		cfg.Stealth.OperatingHoursStart,
		cfg.Stealth.OperatingHoursEnd,
		logger,
	)

	// Check operating hours (temporarily disabled for testing)
	// if !stealthMgr.IsWithinOperatingHours() {
	// 	logger.Warn("Current time is outside configured operating hours")
	// 	logger.Info("Waiting for operating hours...")
	// 	stealthMgr.WaitForOperatingHours()
	// }

	// Setup browser with stealth settings
	logger.Info("Launching browser...")
	browser, err := stealth.SetupStealthBrowser(cfg.Browser.Headless, cfg.Browser.SlowMotion)
	if err != nil {
		logger.Fatalf("Failed to setup browser: %v", err)
	}
	defer browser.Close()

	// Create new page
	page := browser.MustPage()
	defer page.Close()

	// Apply stealth settings to page
	logger.Debug("Applying stealth settings...")
	if err := stealth.ApplyStealthSettings(page); err != nil {
		logger.Fatalf("Failed to apply stealth settings: %v", err)
	}

	// Set random viewport
	if err := stealth.RandomViewport(page); err != nil {
		logger.Warnf("Failed to set random viewport: %v", err)
	}

	// Set random user agent
	if err := stealth.SetRandomUserAgent(page); err != nil {
		logger.Warnf("Failed to set random user agent: %v", err)
	}

	// Initialize authenticator
	authenticator := auth.NewAuthenticator(
		cfg.LinkedIn.Email,
		cfg.LinkedIn.Password,
		cfg.Session.Path,
		logger,
	)

	// Clear session if requested
	if *clearSession {
		logger.Info("Clearing saved session...")
		if err := authenticator.ClearSession(); err != nil {
			logger.Warnf("Failed to clear session: %v", err)
		}
	}

	// Login to LinkedIn (manual login)
	logger.Info("Logging in to LinkedIn...")
	if err := authenticator.ManualLogin(page); err != nil {
		logger.Fatalf("Failed to login: %v", err)
	}

	// Apply anti-detection AFTER login (when we have a real page context)
	logger.Info("Applying advanced anti-detection...")
	fingerprintMgr := stealth.NewFingerprintManager()
	if err := fingerprintMgr.ApplyAllFingerprinting(page); err != nil {
		logger.Warnf("Failed to apply fingerprinting protection: %v", err)
	} else {
		logger.Info("✓ Advanced fingerprinting protection applied")
	}

	// Apply ultra stealth (maximum anti-detection)
	logger.Info("Applying ultra stealth mode...")
	ultraStealth := stealth.NewUltraStealthManager()
	if err := ultraStealth.ApplyUltraStealth(page); err != nil {
		logger.Warnf("Failed to apply ultra stealth: %v", err)
	} else {
		logger.Info("✓ Ultra stealth mode activated")
	}

	// Apply human behavior patterns
	logger.Info("Applying human behavior simulation...")
	if err := stealth.ApplyAllBehaviorPatterns(page); err != nil {
		logger.Warnf("Failed to apply behavior patterns: %v", err)
	} else {
		logger.Info("✓ Human behavior simulation active")
	}
	if err := stealth.InjectBehaviorScripts(page); err != nil {
		logger.Warnf("Failed to inject behavior scripts: %v", err)
	}

	// Initialize managers
	searcher := search.NewSearcher(stealthMgr, db, logger)
	connMgr := connection.NewManager(
		stealthMgr,
		db,
		cfg.Connections.MaxPerDay,
		cfg.Connections.MaxPerHour,
		cfg.Messaging.DefaultMessageTemplate,
		logger,
	)
	msgMgr := messaging.NewManager(
		stealthMgr,
		db,
		cfg.Messaging.MaxPerDay,
		logger,
	)

	// Run interactive menu or execute single action
	if *action == "" && *interactive {
		runInteractiveMode(page, searcher, connMgr, msgMgr, cfg, logger)
	} else {
		// Perform single action based on command
		switch *action {
		case "search":
			performSearch(page, searcher, cfg, logger)
		case "connect":
			performConnect(page, searcher, connMgr, cfg, logger)
		case "message":
			performMessaging(page, msgMgr, logger)
		case "detect-connections":
			performDetectConnections(page, msgMgr, logger)
		default:
			logger.Fatalf("Unknown action: %s. Use -interactive flag or specify -action", *action)
		}
	}

	logger.Info("SubSpace completed successfully!")
}

func runInteractiveMode(page *rod.Page, searcher *search.Searcher, connMgr *connection.Manager, msgMgr *messaging.Manager, cfg *config.Config, logger *logrus.Logger) {
	for {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("           🚀 SubSpace LinkedIn Automation")
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println("\nWhat would you like to do?")
		fmt.Println("\n1. 🔍 Search for people")
		fmt.Println("2. 🤝 Search and send connection requests")
		fmt.Println("3. 💬 Send messages to new connections")
		fmt.Println("4. 🔁 Send auto follow-up to replies")
		fmt.Println("5. 👥 View new connections")
		fmt.Println("6. 📊 View activity statistics")
		fmt.Println("7. ⚙️  Change settings")
		fmt.Println("8. 🔥 Account warm-up session")
		fmt.Println("9. 🧹 Clear session (fixes anti-bot blocks)")
		fmt.Println("10. 🚪 Exit")
		fmt.Print("\nEnter your choice (1-10): ")
		choice := readInput()

		switch choice {
		case "1":
			interactiveSearch(page, searcher, cfg, logger)
		case "2":
			interactiveConnect(page, searcher, connMgr, cfg, logger)
		case "3":
			interactiveMessage(page, msgMgr, logger)
		case "4":
			interactiveFollowUp(page, msgMgr, logger)
		case "5":
			performDetectConnections(page, msgMgr, logger)
			fmt.Print("\nPress Enter to continue...")
			readInput()
		case "6":
			showStatistics(cfg, logger)
		case "7":
			changeSettings(cfg, logger)
		case "8":
			performWarmUp(page, logger)
		case "9":
			performClearSession(page, cfg, logger)
		case "10":
			fmt.Println("\n👋 Thank you for using SubSpace! Goodbye!")
			return
		default:
			fmt.Println("\n❌ Invalid choice. Please enter a number between 1 and 9.")
			fmt.Print("Press Enter to continue...")
			readInput()
		}
	}
}

func interactiveSearch(page *rod.Page, searcher *search.Searcher, cfg *config.Config, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("🔍 SEARCH FOR PEOPLE")
	fmt.Println(strings.Repeat("-", 60))

	// Get keywords
	fmt.Printf("\nEnter search keywords (e.g., Software Engineer, Product Manager)\n[Default: %s]: ", strings.Join(cfg.Search.DefaultKeywords, ", "))
	keywordsInput := readInput()

	searchKeywords := cfg.Search.DefaultKeywords
	if keywordsInput != "" {
		searchKeywords = strings.Split(keywordsInput, ",")
		for i := range searchKeywords {
			searchKeywords[i] = strings.TrimSpace(searchKeywords[i])
		}
	}

	// Get location
	fmt.Printf("Enter location [Default: %s]: ", cfg.Search.DefaultLocation)
	locationInput := readInput()

	searchLocation := cfg.Search.DefaultLocation
	if locationInput != "" {
		searchLocation = locationInput
	}

	// Get max results
	fmt.Print("Enter maximum results [Default: 10]: ")
	maxResultsInput := readInput()

	maxResults := 10
	if maxResultsInput != "" {
		if val, err := strconv.Atoi(maxResultsInput); err == nil {
			maxResults = val
		}
	}

	fmt.Printf("\n🔎 Searching for: %v in %s (max %d results)\n", searchKeywords, searchLocation, maxResults)

	// Perform search
	results, err := searcher.SearchPeople(page, searchKeywords, searchLocation, maxResults)
	if err != nil {
		logger.Errorf("Search failed: %v", err)
		fmt.Println("\n❌ Search failed. Please try again.")
	} else {
		fmt.Printf("\n✅ Found %d results:\n", len(results))
		for i, result := range results {
			fmt.Printf("\n%d. %s\n", i+1, result.Name)
			fmt.Printf("   Title: %s\n", result.Title)
			fmt.Printf("   Company: %s\n", result.Company)
			fmt.Printf("   Location: %s\n", result.Location)
			fmt.Printf("   Profile: %s\n", result.ProfileURL)
		}
	}

	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func interactiveConnect(page *rod.Page, searcher *search.Searcher, connMgr *connection.Manager, cfg *config.Config, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("🤝 SEND CONNECTION REQUESTS")
	fmt.Println(strings.Repeat("-", 60))

	// Get keywords
	fmt.Printf("\nEnter search keywords (e.g., Software Engineer, Product Manager)\n[Default: %s]: ", strings.Join(cfg.Search.DefaultKeywords, ", "))
	keywordsInput := readInput()

	searchKeywords := cfg.Search.DefaultKeywords
	if keywordsInput != "" {
		searchKeywords = strings.Split(keywordsInput, ",")
		for i := range searchKeywords {
			searchKeywords[i] = strings.TrimSpace(searchKeywords[i])
		}
	}

	// Get location
	fmt.Printf("Enter location [Default: %s]: ", cfg.Search.DefaultLocation)
	locationInput := readInput()

	searchLocation := cfg.Search.DefaultLocation
	if locationInput != "" {
		searchLocation = locationInput
	}

	// Get max results
	fmt.Print("Enter maximum connections to send [Default: 5]: ")
	maxResultsInput := readInput()

	maxResults := 5
	if maxResultsInput != "" {
		if val, err := strconv.Atoi(maxResultsInput); err == nil {
			maxResults = val
		}
	}

	// Get custom message
	fmt.Printf("\nEnter connection message (or press Enter for default)\n[Default: %s]: ", cfg.Messaging.DefaultMessageTemplate)
	customMessage := readInput()

	connectionMessage := cfg.Messaging.DefaultMessageTemplate
	if customMessage != "" {
		connectionMessage = customMessage
	}

	fmt.Printf("\n🔎 Searching for people to connect with...\n")

	// Perform search
	results, err := searcher.SearchPeople(page, searchKeywords, searchLocation, maxResults)
	if err != nil {
		logger.Errorf("Search failed: %v", err)
		fmt.Println("\n❌ Search failed. Please try again.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	if len(results) == 0 {
		fmt.Println("\n⚠️  No results found.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	fmt.Printf("\n✅ Found %d people. Preparing to send connection requests...\n", len(results))

	// Confirm before sending
	fmt.Print("\nAre you sure you want to send connection requests? (yes/no): ")
	confirm := readInput()

	if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
		fmt.Println("\n❌ Connection requests cancelled.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	// Send connection requests
	fmt.Println("\n📤 Sending connection requests...")
	successCount, err := connMgr.SendBulkConnectionRequests(page, results, connectionMessage)
	if err != nil {
		logger.Warnf("Some connection requests failed: %v", err)
	}

	fmt.Printf("\n✅ Successfully sent %d out of %d connection requests!\n", successCount, len(results))
	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func interactiveMessage(page *rod.Page, msgMgr *messaging.Manager, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("💬 SEND MESSAGES TO NEW CONNECTIONS")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("\n🔍 Detecting new connections...")

	newConnections, err := msgMgr.DetectNewConnections(page)
	if err != nil {
		logger.Errorf("Failed to detect new connections: %v", err)
		fmt.Println("\n❌ Failed to detect new connections. Please try again.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	if len(newConnections) == 0 {
		fmt.Println("\n✅ No new connections found.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	fmt.Printf("\n✅ Found %d new connections:\n", len(newConnections))
	for i, conn := range newConnections {
		fmt.Printf("%d. %s at %s\n", i+1, conn.ProfileName, conn.Company)
	}

	// Get custom message
	fmt.Print("\nEnter your message (or press Enter for default): ")
	customMessage := readInput()

	if customMessage == "" {
		customMessage = "Thank you for connecting! I'd love to learn more about your work."
	}

	// Confirm before sending
	fmt.Print("\nAre you sure you want to send messages? (yes/no): ")
	confirm := readInput()

	if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
		fmt.Println("\n❌ Messaging cancelled.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	// Prepare recipients
	recipients := make([]struct {
		ProfileURL  string
		ProfileName string
		Message     string
	}, len(newConnections))

	for i, conn := range newConnections {
		recipients[i] = struct {
			ProfileURL  string
			ProfileName string
			Message     string
		}{
			ProfileURL:  conn.ProfileURL,
			ProfileName: conn.ProfileName,
			Message:     customMessage,
		}
	}

	// Send messages
	fmt.Println("\n📤 Sending messages...")
	successCount, err := msgMgr.SendBulkMessages(page, recipients)
	if err != nil {
		logger.Warnf("Some messages failed: %v", err)
	}

	fmt.Printf("\n✅ Successfully sent %d out of %d messages!\n", successCount, len(recipients))
	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func interactiveFollowUp(page *rod.Page, msgMgr *messaging.Manager, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("🔁 AUTO FOLLOW-UP TO REPLIES")
	fmt.Println(strings.Repeat("-", 60))

	// Step 1: Check for new replies
	fmt.Println("\n🔍 Checking for new replies in your inbox...")
	err := msgMgr.CheckForReplies(page)
	if err != nil {
		logger.Errorf("Failed to check for replies: %v", err)
		fmt.Println("\n❌ Failed to check for replies. Please try again.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	// Step 2: Get conversations needing follow-up
	fmt.Println("\n📋 Looking for conversations needing follow-up...")
	time.Sleep(1 * time.Second)

	// Ask for follow-up message
	fmt.Print("\nEnter your follow-up message (use {firstName} or {name} for personalization): ")
	fmt.Println("\nExample: Thanks for your reply, {firstName}! I'd be happy to discuss further.")
	fmt.Print("\nYour message: ")
	followUpMsg := readInput()

	if followUpMsg == "" {
		followUpMsg = "Thanks for your reply, {firstName}! I appreciate you taking the time to respond."
	}

	// Confirm before sending
	fmt.Print("\nAre you sure you want to send follow-up messages? (yes/no): ")
	confirm := readInput()

	if strings.ToLower(confirm) != "yes" && strings.ToLower(confirm) != "y" {
		fmt.Println("\n❌ Follow-up cancelled.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	// Send follow-ups
	fmt.Println("\n📤 Sending follow-up messages...")
	successCount, err := msgMgr.SendFollowUpMessages(page, followUpMsg)
	if err != nil {
		logger.Errorf("Failed to send follow-ups: %v", err)
	}

	if successCount == 0 {
		fmt.Println("\n✅ No conversations need follow-up at this time.")
	} else {
		fmt.Printf("\n✅ Successfully sent %d follow-up message(s)!\n", successCount)
	}

	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func showStatistics(cfg *config.Config, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("📊 ACTIVITY STATISTICS")
	fmt.Println(strings.Repeat("-", 60))

	// Initialize database to read stats
	db, err := storage.NewDB(cfg.Database.Path)
	if err != nil {
		fmt.Printf("\n❌ Failed to load statistics: %v\n", err)
		fmt.Print("\nPress Enter to continue...")
		fmt.Scanln()
		return
	}
	defer db.Close()

	fmt.Println("\n📈 Your LinkedIn automation activity:")
	fmt.Println("\n(Statistics feature coming soon - database integration needed)")
	fmt.Println("This will show:")
	fmt.Println("  • Total searches performed")
	fmt.Println("  • Connection requests sent")
	fmt.Println("  • Messages sent")
	fmt.Println("  • Success rates")

	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func changeSettings(cfg *config.Config, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("⚙️  SETTINGS")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("\nCurrent Settings:")
	fmt.Printf("  • Default Keywords: %s\n", strings.Join(cfg.Search.DefaultKeywords, ", "))
	fmt.Printf("  • Default Location: %s\n", cfg.Search.DefaultLocation)
	fmt.Printf("  • Max Connections/Day: %d\n", cfg.Connections.MaxPerDay)
	fmt.Printf("  • Max Messages/Day: %d\n", cfg.Messaging.MaxPerDay)
	fmt.Printf("  • Operating Hours: %d:00 - %d:00\n", cfg.Stealth.OperatingHoursStart, cfg.Stealth.OperatingHoursEnd)

	fmt.Println("\n💡 To change settings, edit the config.yaml file or .env file")
	fmt.Println("   and restart the application.")

	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func performSearch(page *rod.Page, searcher *search.Searcher, cfg *config.Config, logger *logrus.Logger) {
	// Determine search keywords
	searchKeywords := cfg.Search.DefaultKeywords
	if *keywords != "" {
		searchKeywords = strings.Split(*keywords, ",")
		for i := range searchKeywords {
			searchKeywords[i] = strings.TrimSpace(searchKeywords[i])
		}
	}

	// Determine location
	searchLocation := cfg.Search.DefaultLocation
	if *location != "" {
		searchLocation = *location
	}

	logger.Infof("Searching for: %v in %s", searchKeywords, searchLocation)

	// Perform search
	results, err := searcher.SearchPeople(page, searchKeywords, searchLocation, *maxResults)
	if err != nil {
		logger.Fatalf("Search failed: %v", err)
	}

	// Display results
	logger.Infof("Found %d results:", len(results))
	for i, result := range results {
		logger.Infof("%d. %s - %s at %s", i+1, result.Name, result.Title, result.Company)
		logger.Infof("   Profile: %s", result.ProfileURL)
	}
}

func performConnect(page *rod.Page, searcher *search.Searcher, connMgr *connection.Manager, cfg *config.Config, logger *logrus.Logger) {
	// First, perform search
	searchKeywords := cfg.Search.DefaultKeywords
	if *keywords != "" {
		searchKeywords = strings.Split(*keywords, ",")
		for i := range searchKeywords {
			searchKeywords[i] = strings.TrimSpace(searchKeywords[i])
		}
	}

	searchLocation := cfg.Search.DefaultLocation
	if *location != "" {
		searchLocation = *location
	}

	logger.Infof("Searching for people to connect with...")
	results, err := searcher.SearchPeople(page, searchKeywords, searchLocation, *maxResults)
	if err != nil {
		logger.Fatalf("Search failed: %v", err)
	}

	// Determine message to use
	connectionMessage := cfg.Messaging.DefaultMessageTemplate
	if *message != "" {
		connectionMessage = *message
	}

	// Send connection requests
	logger.Infof("Sending connection requests to %d people...", len(results))
	successCount, err := connMgr.SendBulkConnectionRequests(page, results, connectionMessage)
	if err != nil {
		logger.Errorf("Bulk connection requests encountered errors: %v", err)
	}

	logger.Infof("Successfully sent %d connection requests", successCount)
}

func performMessaging(page *rod.Page, msgMgr *messaging.Manager, logger *logrus.Logger) {
	// Detect new connections
	logger.Info("Detecting new connections...")
	newConnections, err := msgMgr.DetectNewConnections(page)
	if err != nil {
		logger.Fatalf("Failed to detect new connections: %v", err)
	}

	if len(newConnections) == 0 {
		logger.Info("No new connections found")
		return
	}

	// Prepare messages
	messageTemplate := *message
	if messageTemplate == "" {
		messageTemplate = "Thank you for connecting! I'd love to learn more about your work."
	}

	recipients := make([]struct {
		ProfileURL  string
		ProfileName string
		Message     string
	}, len(newConnections))

	for i, conn := range newConnections {
		recipients[i] = struct {
			ProfileURL  string
			ProfileName string
			Message     string
		}{
			ProfileURL:  conn.ProfileURL,
			ProfileName: conn.ProfileName,
			Message:     messageTemplate,
		}
	}

	// Send messages
	logger.Infof("Sending follow-up messages to %d new connections...", len(recipients))
	successCount, err := msgMgr.SendBulkMessages(page, recipients)
	if err != nil {
		logger.Errorf("Bulk messaging encountered errors: %v", err)
	}

	logger.Infof("Successfully sent %d messages", successCount)
}

func performDetectConnections(page *rod.Page, msgMgr *messaging.Manager, logger *logrus.Logger) {
	logger.Info("Detecting new connections...")
	newConnections, err := msgMgr.DetectNewConnections(page)
	if err != nil {
		logger.Fatalf("Failed to detect new connections: %v", err)
	}

	logger.Infof("Found %d new connections:", len(newConnections))
	for i, conn := range newConnections {
		logger.Infof("%d. %s at %s", i+1, conn.ProfileName, conn.Company)
		logger.Infof("   Profile: %s", conn.ProfileURL)
	}
}

func performWarmUp(page *rod.Page, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("🔥 ACCOUNT WARM-UP SESSION")
	fmt.Println(strings.Repeat("-", 60))

	sessionBuilder := stealth.NewSessionBuilder(logger)

	fmt.Println("\n📋 Recommended Warm-Up Plan:")
	fmt.Println(sessionBuilder.RecommendedWarmUpPlan())

	fmt.Print("\nHow long would you like to warm up? (minutes) [Default: 10]: ")
	durationInput := readInput()

	duration := 10
	if durationInput != "" {
		if val, err := strconv.Atoi(durationInput); err == nil {
			duration = val
		}
	}

	fmt.Printf("\n🔥 Starting %d-minute warm-up session...\n", duration)
	fmt.Println("The bot will perform natural browsing activities.")
	fmt.Println("Press Ctrl+C to stop early.\n")

	if err := sessionBuilder.WarmUpSession(page, time.Duration(duration)*time.Minute); err != nil {
		logger.Errorf("Warm-up failed: %v", err)
		fmt.Println("\n❌ Warm-up session encountered errors.")
	} else {
		fmt.Println("\n✅ Warm-up session completed successfully!")
		fmt.Printf("Total actions performed: %d\n", duration/2) // Approximate
	}

	fmt.Print("\nPress Enter to continue...")
	readInput()
}

func performClearSession(page *rod.Page, cfg *config.Config, logger *logrus.Logger) {
	fmt.Println("\n" + strings.Repeat("-", 60))
	fmt.Println("🧹 CLEAR SESSION & RESET TRUST")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Println("\n⚠️  This will:")
	fmt.Println("   1. Clear all browser cookies")
	fmt.Println("   2. Clear localStorage")
	fmt.Println("   3. Delete saved session file")
	fmt.Println("   4. Require you to login again manually")
	fmt.Println("\n💡 Use this if you're seeing all skeleton/blank profiles")
	fmt.Println("   (LinkedIn anti-bot detection is blocking real data)")

	fmt.Print("\nAre you sure you want to clear the session? (y/N): ")
	confirm := readInput()

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("\n❌ Cancelled. Session not cleared.")
		fmt.Print("\nPress Enter to continue...")
		readInput()
		return
	}

	// Clear browser data
	fmt.Println("\n🗑️  Clearing browser cookies...")
	cookies, err := page.Cookies(nil)
	if err == nil && len(cookies) > 0 {
		page.SetCookies(nil) // Clear all
		logger.Info("Cleared browser cookies")
	}

	// Clear localStorage
	fmt.Println("🗑️  Clearing localStorage...")
	page.Eval(`() => {
		try {
			localStorage.clear();
			sessionStorage.clear();
		} catch(e) {}
	}`)
	logger.Info("Cleared localStorage and sessionStorage")

	// Delete session file
	sessionPath := cfg.Session.Path
	if sessionPath == "" {
		sessionPath = "data/sessions/cookies.json"
	}
	fmt.Printf("🗑️  Deleting session file: %s\n", sessionPath)
	if err := os.Remove(sessionPath); err != nil {
		if !os.IsNotExist(err) {
			logger.Warnf("Could not delete session file: %v", err)
		}
	} else {
		logger.Info("Deleted session file")
	}

	fmt.Println("\n✅ Session cleared successfully!")
	fmt.Println("\n⚠️  You need to restart the application and login again.")
	fmt.Println("   After login, run Option 7 (warm-up session) for 10-15 minutes")
	fmt.Println("   to rebuild trust before searching.")

	fmt.Print("\nPress Enter to exit and restart...")
	readInput()

	// Exit so user can restart fresh
	fmt.Println("\n👋 Exiting... Please restart the application.")
	os.Exit(0)
}

func setupLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}
