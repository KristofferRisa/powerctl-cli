package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

const (
	// GraphQLEndpoint is the Tibber GraphQL API URL
	GraphQLEndpoint = "https://api.tibber.com/v1-beta/gql"

	// DefaultTimeout for HTTP requests
	DefaultTimeout = 30 * time.Second

	// MaxConsumptionPage is the most Nodes Tibber returns for one consumption request
	MaxConsumptionPage = 744
)

// Client handles communication with Tibber API
type Client struct {
	token      string
	httpClient *http.Client
	endpoint   string
	now        func() time.Time
}

// NewClient creates a new Tibber API client
func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		endpoint: GraphQLEndpoint,
		now:      time.Now,
	}
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// execute sends a GraphQL request and returns the raw response
func (c *Client) execute(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}

// GetHomes fetches all homes for the authenticated user
func (c *Client) GetHomes(ctx context.Context) ([]models.HomeResponse, error) {
	data, err := c.execute(ctx, QueryHomes, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Viewer struct {
			Homes []models.HomeResponse `json:"homes"`
		} `json:"viewer"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse homes: %w", err)
	}

	return result.Viewer.Homes, nil
}

// GetPrices fetches price information for a specific home
func (c *Client) GetPrices(ctx context.Context, homeID string) (*models.PriceInfo, error) {
	variables := map[string]interface{}{}

	data, err := c.execute(ctx, QueryPrices, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Viewer struct {
			Homes []struct {
				ID                  string `json:"id"`
				CurrentSubscription *struct {
					PriceInfo *models.PriceInfo `json:"priceInfo"`
				} `json:"currentSubscription"`
			} `json:"homes"`
		} `json:"viewer"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse prices: %w", err)
	}

	// Find the matching home or use first one
	for _, home := range result.Viewer.Homes {
		if homeID == "" || home.ID == homeID {
			if home.CurrentSubscription != nil && home.CurrentSubscription.PriceInfo != nil {
				return home.CurrentSubscription.PriceInfo, nil
			}
		}
	}

	return nil, fmt.Errorf("no price information found")
}

// ConsumptionRange bounds a Consumption History fetch. Set either Last, or From
// with an optional To. From is inclusive and To is exclusive; a zero To means
// up to now. Every Node that overlaps the range is returned, so with a coarse
// Resolution the first and last Node may extend past From and To.
type ConsumptionRange struct {
	Last int
	From time.Time
	To   time.Time
}

// GetConsumptionHistory fetches historical consumption data, oldest Node first.
// Tibber returns at most MaxConsumptionPage Nodes per request, so longer ranges
// are fetched page by page, backwards from now, or from To when it is more than
// one page in the past. short reports that Tibber's history ended before the
// requested range was covered.
func (c *Client) GetConsumptionHistory(ctx context.Context, homeID string, resolution string, r ConsumptionRange) (nodes []models.ConsumptionNode, short bool, err error) {
	before := c.cursorBefore(resolution, r.To)
	fromBuiltCursor := before != ""
	for {
		size := MaxConsumptionPage
		if r.Last > 0 && r.Last-len(nodes) < size {
			size = r.Last - len(nodes)
		}

		page, err := c.getConsumptionPage(ctx, homeID, resolution, size, before)
		if fromBuiltCursor {
			fromBuiltCursor = false
			// Tibber may have changed its cursor format, so if the page is not
			// the one that ends at To, page from now instead
			if err != nil || len(page.Nodes) == 0 || page.Nodes[len(page.Nodes)-1].To.Before(r.To) {
				before = ""
				continue
			}
		}
		if err != nil {
			return nil, false, err
		}
		fetched := len(nodes)
		nodes = prependOlder(page.Nodes, nodes)

		if r.Last > 0 && len(nodes) >= r.Last {
			break
		}
		if r.Last == 0 && len(nodes) > 0 && !nodes[0].From.After(r.From) {
			break
		}
		// A page with no older Nodes means the history has ended, and also
		// stops a cursor that never moves from looping forever.
		if !page.PageInfo.HasPreviousPage || len(nodes) == fetched {
			short = true
			break
		}
		before = page.PageInfo.StartCursor
	}

	if r.Last == 0 {
		nodes = trimToRange(nodes, r.From, r.To)
	}
	return nodes, short, nil
}

// longestNode is the most time one Node can cover at each Resolution
var longestNode = map[string]time.Duration{
	"HOURLY":  time.Hour,
	"DAILY":   25 * time.Hour, // a day that leaves daylight saving time
	"WEEKLY":  7*24*time.Hour + time.Hour,
	"MONTHLY": 31*24*time.Hour + time.Hour,
	"ANNUAL":  366*24*time.Hour + time.Hour,
}

// pageSize is the most Nodes Tibber returns for one request at each
// Resolution, whatever last asks for. ANNUAL has not been measured.
var pageSize = map[string]int{
	"HOURLY":  MaxConsumptionPage,
	"DAILY":   31,
	"WEEKLY":  52,
	"MONTHLY": 12,
	"ANNUAL":  MaxConsumptionPage,
}

// cursorBefore builds the cursor Tibber would give a Node starting at to, so
// a range in the past can start paging at to instead of now. The cursor format
// is undocumented (it is the base64 of a Node's from timestamp), so it is only
// built when to is more than one page in the past, where it saves a request.
func (c *Client) cursorBefore(resolution string, to time.Time) string {
	span, ok := longestNode[resolution]
	if to.IsZero() || !ok {
		return ""
	}
	// Tibber reads a cursor as the start of the Node it falls in, so a to
	// inside an hour moves to the end of it to keep that Node in the page
	if resolution == "HOURLY" {
		if start := to.Truncate(time.Hour); start.Before(to) {
			to = start.Add(time.Hour)
		}
	}
	if !to.Before(c.now().Add(-time.Duration(pageSize[resolution]) * span)) {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(to.Format("2006-01-02T15:04:05.000-07:00")))
}

// prependOlder puts page in front of nodes, dropping any Node in page that is
// not older than the oldest Node already fetched, so a page boundary never
// duplicates a Node.
func prependOlder(page, nodes []models.ConsumptionNode) []models.ConsumptionNode {
	if len(nodes) == 0 {
		return page
	}
	oldest := nodes[0].From
	merged := make([]models.ConsumptionNode, 0, len(page)+len(nodes))
	for _, n := range page {
		if n.From.Before(oldest) {
			merged = append(merged, n)
		}
	}
	return append(merged, nodes...)
}

// trimToRange keeps the Nodes that overlap [from, to), so a period that is only
// partly inside the range is kept at both ends. A zero to has no upper bound.
func trimToRange(nodes []models.ConsumptionNode, from, to time.Time) []models.ConsumptionNode {
	trimmed := make([]models.ConsumptionNode, 0, len(nodes))
	for _, n := range nodes {
		if !n.To.After(from) || (!to.IsZero() && !n.From.Before(to)) {
			continue
		}
		trimmed = append(trimmed, n)
	}
	return trimmed
}

// getConsumptionPage fetches the last Nodes before the cursor, or the newest
// Nodes when before is empty
func (c *Client) getConsumptionPage(ctx context.Context, homeID string, resolution string, last int, before string) (*models.ConsumptionHistory, error) {
	variables := map[string]interface{}{
		"resolution": resolution,
		"last":       last,
	}
	if before != "" {
		variables["before"] = before
	}

	var data json.RawMessage
	var err error

	if homeID != "" {
		variables["homeId"] = homeID
		data, err = c.execute(ctx, QueryConsumptionHome, variables)
	} else {
		data, err = c.execute(ctx, QueryConsumptionHomes, variables)
	}

	if err != nil {
		if homeID != "" && strings.Contains(err.Error(), "does not exist") {
			return nil, fmt.Errorf("home with ID %q not found: %w", homeID, err)
		}
		return nil, err
	}

	type graphQLResponse struct {
		Viewer struct {
			Home *struct {
				ID          string                    `json:"id"`
				Consumption models.ConsumptionHistory `json:"consumption"`
			} `json:"home"`
			Homes []struct {
				ID          string                    `json:"id"`
				Consumption models.ConsumptionHistory `json:"consumption"`
			} `json:"homes"`
		} `json:"viewer"`
	}

	var result graphQLResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse consumption history: %w", err)
	}

	if homeID != "" && result.Viewer.Home != nil {
		return &result.Viewer.Home.Consumption, nil
	}

	if homeID != "" {
		return nil, fmt.Errorf("home with ID %q not found", homeID)
	}

	if len(result.Viewer.Homes) == 0 {
		return nil, fmt.Errorf("no homes found")
	}

	// Return first home's consumption
	return &result.Viewer.Homes[0].Consumption, nil
}

// GetTimeZone fetches the time zone of a home, or of the first home when
// homeID is empty
func (c *Client) GetTimeZone(ctx context.Context, homeID string) (*time.Location, error) {
	var data json.RawMessage
	var err error

	if homeID != "" {
		data, err = c.execute(ctx, QueryTimeZoneHome, map[string]interface{}{"homeId": homeID})
	} else {
		data, err = c.execute(ctx, QueryTimeZoneHomes, nil)
	}
	if err != nil {
		return nil, err
	}

	type home struct {
		TimeZone string `json:"timeZone"`
	}
	var result struct {
		Viewer struct {
			Home  *home  `json:"home"`
			Homes []home `json:"homes"`
		} `json:"viewer"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse time zone: %w", err)
	}

	var name string
	switch {
	case homeID != "" && result.Viewer.Home != nil:
		name = result.Viewer.Home.TimeZone
	case homeID == "" && len(result.Viewer.Homes) > 0:
		name = result.Viewer.Homes[0].TimeZone
	default:
		return nil, fmt.Errorf("no home found to read the time zone from")
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("unknown home time zone %q: %w", name, err)
	}
	return loc, nil
}
