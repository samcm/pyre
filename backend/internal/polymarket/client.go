package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	baseURL        = "https://data-api.polymarket.com"
	defaultTimeout = 30 * time.Second
)

// Client defines the interface for Polymarket API operations
type Client interface {
	GetPositions(ctx context.Context, address string) (PositionsResponse, error)
	GetTrades(ctx context.Context, address string, limit int) (TradesResponse, error)
	GetActivity(ctx context.Context, address string) (ActivitiesResponse, error)
	GetUserProfile(ctx context.Context, address string) (*ProfileResponse, error)
	GetPortfolioStats(ctx context.Context, username string, address string) (*PortfolioStats, error)
}

// client implements the Polymarket API client
type client struct {
	httpClient *http.Client
	baseURL    string
	log        logrus.FieldLogger
}

var _ Client = (*client)(nil)

// NewClient creates a new Polymarket API client
func NewClient(log logrus.FieldLogger) Client {
	return &client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
		log:     log.WithField("package", "polymarket"),
	}
}

// GetPositions fetches positions for a given address
func (c *client) GetPositions(ctx context.Context, address string) (PositionsResponse, error) {
	c.log.WithField("address", address).Debug("fetching positions")

	endpoint := fmt.Sprintf("%s/positions", c.baseURL)
	params := url.Values{}
	params.Add("user", address)

	var positions PositionsResponse
	if err := c.doRequest(ctx, endpoint, params, &positions); err != nil {
		return nil, fmt.Errorf("failed to fetch positions for %s: %w", address, err)
	}

	c.log.WithFields(logrus.Fields{
		"address": address,
		"count":   len(positions),
	}).Debug("fetched positions")

	return positions, nil
}

// GetTrades fetches trades for a given address
func (c *client) GetTrades(ctx context.Context, address string, limit int) (TradesResponse, error) {
	c.log.WithFields(logrus.Fields{
		"address": address,
		"limit":   limit,
	}).Debug("fetching trades")

	endpoint := fmt.Sprintf("%s/trades", c.baseURL)
	params := url.Values{}
	params.Add("user", address)
	if limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", limit))
	}

	var trades TradesResponse
	if err := c.doRequest(ctx, endpoint, params, &trades); err != nil {
		return nil, fmt.Errorf("failed to fetch trades for %s: %w", address, err)
	}

	c.log.WithFields(logrus.Fields{
		"address": address,
		"count":   len(trades),
	}).Debug("fetched trades")

	return trades, nil
}

// GetActivity fetches activity for a given address
func (c *client) GetActivity(ctx context.Context, address string) (ActivitiesResponse, error) {
	c.log.WithField("address", address).Debug("fetching activity")

	endpoint := fmt.Sprintf("%s/activity", c.baseURL)
	params := url.Values{}
	params.Add("user", address)

	var activities ActivitiesResponse
	if err := c.doRequest(ctx, endpoint, params, &activities); err != nil {
		return nil, fmt.Errorf("failed to fetch activity for %s: %w", address, err)
	}

	c.log.WithFields(logrus.Fields{
		"address": address,
		"count":   len(activities),
	}).Debug("fetched activity")

	return activities, nil
}

// GetUserProfile fetches profile data for a given address
func (c *client) GetUserProfile(ctx context.Context, address string) (*ProfileResponse, error) {
	c.log.WithField("address", address).Debug("fetching user profile")

	endpoint := fmt.Sprintf("%s/activity", c.baseURL)
	params := url.Values{}
	params.Add("user", address)
	params.Add("limit", "1")

	// The activity endpoint returns profile data embedded in each activity
	var activities []struct {
		ProfileResponse
	}
	if err := c.doRequest(ctx, endpoint, params, &activities); err != nil {
		return nil, fmt.Errorf("failed to fetch profile for %s: %w", address, err)
	}

	if len(activities) == 0 {
		return nil, nil // No activity, no profile data available
	}

	profile := &ProfileResponse{
		Name:                  activities[0].Name,
		Pseudonym:             activities[0].Pseudonym,
		Bio:                   activities[0].Bio,
		ProfileImage:          activities[0].ProfileImage,
		ProfileImageOptimized: activities[0].ProfileImageOptimized,
	}

	c.log.WithFields(logrus.Fields{
		"address": address,
		"name":    profile.Name,
	}).Debug("fetched user profile")

	return profile, nil
}

// doRequest performs an HTTP GET request and unmarshals the response
func (c *client) doRequest(ctx context.Context, endpoint string, params url.Values, result any) error {
	// Build URL with query parameters
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse endpoint URL: %w", err)
	}
	u.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "pyre/1.0")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Unmarshal response
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// GetPortfolioStats fetches the all-time portfolio stats from Polymarket's profile page
// This scrapes the embedded JSON data since the data API doesn't expose historical PnL
func (c *client) GetPortfolioStats(ctx context.Context, username, address string) (*PortfolioStats, error) {
	c.log.WithFields(logrus.Fields{
		"username": username,
		"address":  address,
	}).Debug("fetching portfolio stats from profile page")

	// Fetch the profile page HTML
	profileURL := fmt.Sprintf("https://polymarket.com/profile/@%s", username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create profile request: %w", err)
	}

	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; pyre/1.0)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile page: %w", err)
	}

	htmlContent := string(body)

	// The PnL data is embedded in the page as part of React Query dehydrated state
	// Look for the pattern: "amount":NUMBER,"pnl":NUMBER
	pnlPattern := regexp.MustCompile(`"amount":([\d.-]+),"pnl":([\d.-]+)`)
	matches := pnlPattern.FindStringSubmatch(htmlContent)

	if len(matches) < 3 {
		// Try alternative pattern - look for positions value data
		// Pattern: ["positions","value","ADDRESS"]
		positionsPattern := regexp.MustCompile(
			`\["positions","value","` + strings.ToLower(address) + `"\][^{]*\{[^}]*"pnl":([\d.-]+)`,
		)
		altMatches := positionsPattern.FindStringSubmatch(htmlContent)
		if len(altMatches) < 2 {
			c.log.WithField("username", username).Warn("could not find PnL data in profile page")
			return nil, nil
		}

		var pnl float64
		if _, err := fmt.Sscanf(altMatches[1], "%f", &pnl); err != nil {
			return nil, fmt.Errorf("failed to parse PnL: %w", err)
		}

		return &PortfolioStats{
			TotalPnl: pnl,
		}, nil
	}

	var amount, pnl float64
	if _, err := fmt.Sscanf(matches[1], "%f", &amount); err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}
	if _, err := fmt.Sscanf(matches[2], "%f", &pnl); err != nil {
		return nil, fmt.Errorf("failed to parse PnL: %w", err)
	}

	stats := &PortfolioStats{
		TotalPnl:    pnl,
		TotalVolume: amount,
	}

	c.log.WithFields(logrus.Fields{
		"username": username,
		"pnl":      stats.TotalPnl,
		"volume":   stats.TotalVolume,
	}).Debug("fetched portfolio stats")

	return stats, nil
}
