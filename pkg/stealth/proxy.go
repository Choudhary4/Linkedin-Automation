package stealth

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ProxyRotator manages residential proxy rotation
type ProxyRotator struct {
	proxies     []Proxy
	currentIdx  int
	mu          sync.RWMutex
	logger      *logrus.Logger
	successRate map[string]float64
	lastUsed    map[string]time.Time
}

// Proxy represents a proxy configuration
type Proxy struct {
	Host     string
	Port     string
	Username string
	Password string
	Country  string
	City     string
	ISP      string
	Type     string // residential, datacenter, mobile
}

// NewProxyRotator creates a new proxy rotator
func NewProxyRotator(proxies []Proxy, logger *logrus.Logger) *ProxyRotator {
	return &ProxyRotator{
		proxies:     proxies,
		currentIdx:  0,
		logger:      logger,
		successRate: make(map[string]float64),
		lastUsed:    make(map[string]time.Time),
	}
}

// GetNextProxy returns the next proxy in rotation with smart selection
func (pr *ProxyRotator) GetNextProxy() *Proxy {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if len(pr.proxies) == 0 {
		return nil
	}

	// Use weighted random selection based on success rate
	bestProxies := pr.selectBestProxies(3)
	if len(bestProxies) == 0 {
		// Fallback to round-robin
		proxy := &pr.proxies[pr.currentIdx]
		pr.currentIdx = (pr.currentIdx + 1) % len(pr.proxies)
		return proxy
	}

	// Randomly select from best proxies
	selected := bestProxies[rand.Intn(len(bestProxies))]
	pr.lastUsed[selected.Host] = time.Now()

	pr.logger.Debugf("Selected proxy: %s (%s, %s)", selected.Host, selected.Country, selected.City)
	return selected
}

// selectBestProxies returns proxies with highest success rate and not recently used
func (pr *ProxyRotator) selectBestProxies(count int) []*Proxy {
	type proxyScore struct {
		proxy *Proxy
		score float64
	}

	var scores []proxyScore
	now := time.Now()

	for i := range pr.proxies {
		proxy := &pr.proxies[i]
		key := proxy.Host

		// Calculate score based on success rate and last used time
		successScore := pr.successRate[key]
		if successScore == 0 {
			successScore = 0.5 // Default for new proxies
		}

		// Prefer proxies not used in last 5 minutes
		timeSinceUse := now.Sub(pr.lastUsed[key]).Minutes()
		timeScore := 1.0
		if timeSinceUse < 5 {
			timeScore = timeSinceUse / 5.0
		}

		// Prefer residential proxies
		typeScore := 1.0
		if proxy.Type == "residential" {
			typeScore = 1.5
		} else if proxy.Type == "mobile" {
			typeScore = 1.3
		}

		totalScore := successScore * timeScore * typeScore
		scores = append(scores, proxyScore{proxy: proxy, score: totalScore})
	}

	// Sort by score (simple bubble sort for small lists)
	for i := 0; i < len(scores); i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Return top N
	result := make([]*Proxy, 0, count)
	for i := 0; i < count && i < len(scores); i++ {
		result = append(result, scores[i].proxy)
	}

	return result
}

// RecordSuccess updates the success rate for a proxy
func (pr *ProxyRotator) RecordSuccess(proxy *Proxy, success bool) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	key := proxy.Host
	currentRate := pr.successRate[key]

	// Exponential moving average
	if success {
		pr.successRate[key] = currentRate*0.9 + 0.1
	} else {
		pr.successRate[key] = currentRate * 0.9
	}

	pr.logger.Debugf("Proxy %s success rate: %.2f", key, pr.successRate[key])
}

// GetProxyURL returns the full proxy URL for rod browser
func (pr *ProxyRotator) GetProxyURL(proxy *Proxy) string {
	if proxy.Username != "" && proxy.Password != "" {
		return fmt.Sprintf("http://%s:%s@%s:%s",
			url.QueryEscape(proxy.Username),
			url.QueryEscape(proxy.Password),
			proxy.Host,
			proxy.Port)
	}
	return fmt.Sprintf("http://%s:%s", proxy.Host, proxy.Port)
}

// TestProxy tests if a proxy is working
func (pr *ProxyRotator) TestProxy(proxy *Proxy) bool {
	proxyURL, err := url.Parse(pr.GetProxyURL(proxy))
	if err != nil {
		pr.logger.Warnf("Invalid proxy URL: %v", err)
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	// Test connection to a reliable endpoint
	resp, err := client.Get("https://www.google.com")
	if err != nil {
		pr.logger.Warnf("Proxy test failed for %s: %v", proxy.Host, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// LoadProxiesFromConfig loads proxies from configuration
func LoadProxiesFromConfig(config []map[string]string) []Proxy {
	proxies := make([]Proxy, 0, len(config))

	for _, p := range config {
		proxy := Proxy{
			Host:     p["host"],
			Port:     p["port"],
			Username: p["username"],
			Password: p["password"],
			Country:  p["country"],
			City:     p["city"],
			ISP:      p["isp"],
			Type:     p["type"],
		}

		if proxy.Type == "" {
			proxy.Type = "residential"
		}

		proxies = append(proxies, proxy)
	}

	return proxies
}

// GetProxyForLocation returns a proxy from a specific location
func (pr *ProxyRotator) GetProxyForLocation(country, city string) *Proxy {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var candidates []*Proxy

	// Find proxies matching location
	for i := range pr.proxies {
		proxy := &pr.proxies[i]
		if proxy.Country == country {
			if city == "" || proxy.City == city {
				candidates = append(candidates, proxy)
			}
		}
	}

	if len(candidates) == 0 {
		pr.logger.Warnf("No proxies found for location %s/%s", country, city)
		return nil
	}

	// Return random from candidates
	return candidates[rand.Intn(len(candidates))]
}

// RotateProxyPeriodically rotates proxy every N minutes
func (pr *ProxyRotator) RotateProxyPeriodically(interval time.Duration, onRotate func(*Proxy)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		proxy := pr.GetNextProxy()
		if proxy != nil && onRotate != nil {
			onRotate(proxy)
		}
	}
}

// GetStatistics returns proxy usage statistics
func (pr *ProxyRotator) GetStatistics() map[string]interface{} {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	stats := map[string]interface{}{
		"total_proxies": len(pr.proxies),
		"success_rates": make(map[string]float64),
		"last_used":     make(map[string]string),
	}

	for host, rate := range pr.successRate {
		stats["success_rates"].(map[string]float64)[host] = rate
	}

	for host, t := range pr.lastUsed {
		stats["last_used"].(map[string]string)[host] = t.Format(time.RFC3339)
	}

	return stats
}

// Example proxy configurations for testing
var ExampleProxies = []Proxy{
	{
		Host:    "proxy1.example.com",
		Port:    "8080",
		Country: "US",
		City:    "New York",
		ISP:     "Verizon",
		Type:    "residential",
	},
	{
		Host:    "proxy2.example.com",
		Port:    "8080",
		Country: "US",
		City:    "Los Angeles",
		ISP:     "AT&T",
		Type:    "residential",
	},
	{
		Host:    "proxy3.example.com",
		Port:    "8080",
		Country: "India",
		City:    "Pune",
		ISP:     "Airtel",
		Type:    "residential",
	},
}
