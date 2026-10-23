package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.token != "test-token" {
		t.Errorf("token = %q, want %q", client.token, "test-token")
	}
	if client.endpoint != GraphQLEndpoint {
		t.Errorf("endpoint = %q, want %q", client.endpoint, GraphQLEndpoint)
	}
}

func TestClient_GetHomes_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want Bearer test-token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
		}

		// Return mock response
		response := `{
			"data": {
				"viewer": {
					"homes": [
						{
							"id": "home-123",
							"appNickname": "Test Home",
							"size": 100,
							"type": "APARTMENT",
							"features": {
								"realTimeConsumptionEnabled": true
							},
							"address": {
								"address1": "123 Test St",
								"postalCode": "12345",
								"city": "Oslo",
								"country": "Norway"
							}
						}
					]
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with mock endpoint
	client := NewClient("test-token")
	client.endpoint = server.URL

	// Test
	homes, err := client.GetHomes(context.Background())
	if err != nil {
		t.Fatalf("GetHomes() error = %v", err)
	}

	if len(homes) != 1 {
		t.Fatalf("len(homes) = %d, want 1", len(homes))
	}

	home := homes[0]
	if home.ID != "home-123" {
		t.Errorf("home.ID = %q, want home-123", home.ID)
	}
	if home.AppNickname != "Test Home" {
		t.Errorf("home.AppNickname = %q, want Test Home", home.AppNickname)
	}
	if !home.Features.RealTimeConsumptionEnabled {
		t.Error("home.Features.RealTimeConsumptionEnabled = false, want true")
	}
}

func TestClient_GetHomes_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": null,
			"errors": [{"message": "Invalid token"}]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient("bad-token")
	client.endpoint = server.URL

	_, err := client.GetHomes(context.Background())
	if err == nil {
		t.Error("GetHomes() should return error for GraphQL error response")
	}
	if err.Error() != "GraphQL error: Invalid token" {
		t.Errorf("error = %q, want 'GraphQL error: Invalid token'", err.Error())
	}
}

func TestClient_GetHomes_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.endpoint = server.URL

	_, err := client.GetHomes(context.Background())
	if err == nil {
		t.Error("GetHomes() should return error for HTTP error")
	}
}

func TestClient_GetPrices_Success(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"viewer": {
					"homes": [
						{
							"id": "home-123",
							"currentSubscription": {
								"priceInfo": {
									"current": {
										"total": 0.45,
										"energy": 0.35,
										"tax": 0.10,
										"startsAt": "` + now + `",
										"level": "NORMAL",
										"currency": "NOK"
									},
									"today": [
										{
											"total": 0.40,
											"energy": 0.30,
											"tax": 0.10,
											"startsAt": "` + now + `",
											"level": "CHEAP",
											"currency": "NOK"
										}
									],
									"tomorrow": []
								}
							}
						}
					]
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.endpoint = server.URL

	prices, err := client.GetPrices(context.Background(), "home-123")
	if err != nil {
		t.Fatalf("GetPrices() error = %v", err)
	}

	if prices.Current == nil {
		t.Fatal("prices.Current is nil")
	}
	if prices.Current.Total != 0.45 {
		t.Errorf("current.Total = %v, want 0.45", prices.Current.Total)
	}
	if prices.Current.Level != "NORMAL" {
		t.Errorf("current.Level = %q, want NORMAL", prices.Current.Level)
	}
	if len(prices.Today) != 1 {
		t.Errorf("len(Today) = %d, want 1", len(prices.Today))
	}
}

func TestClient_GetPrices_NoPriceInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"viewer": {
					"homes": [
						{
							"id": "home-123",
							"currentSubscription": null
						}
					]
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.endpoint = server.URL

	_, err := client.GetPrices(context.Background(), "home-123")
	if err == nil {
		t.Error("GetPrices() should return error when no price info")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.endpoint = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetHomes(ctx)
	if err == nil {
		t.Error("GetHomes() should return error on context timeout")
	}
}

func TestGraphQLRequest_Marshaling(t *testing.T) {
	req := GraphQLRequest{
		Query: "{ viewer { homes { id } } }",
		Variables: map[string]interface{}{
			"homeId": "123",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if result["query"] != req.Query {
		t.Errorf("query = %v, want %v", result["query"], req.Query)
	}
}

func TestClient_GetConsumptionHistory_Success(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"data": {
				"viewer": {
					"home": {
						"id": "home-123",
						"consumption": {
							"nodes": [
								{
									"from": "` + yesterday + `",
									"to": "` + now + `",
									"consumption": 24.5,
									"cost": 120.4,
									"unitPrice": 4.9,
									"unitPriceVAT": 1.22,
									"currency": "NOK"
								},
								{
									"from": "` + now + `",
									"to": "2099-01-01T00:00:00Z",
									"consumption": null,
									"cost": null,
									"unitPrice": null,
									"unitPriceVAT": null,
									"currency": "NOK"
								}
							]
						}
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.endpoint = server.URL

	history, _, err := client.GetConsumptionHistory(context.Background(), "home-123", "DAILY", ConsumptionRange{Last: 2})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("len(history) = %d, want 2", len(history))
	}

	// Check first node (valid data)
	n1 := history[0]
	if n1.Consumption == nil || *n1.Consumption != 24.5 {
		t.Errorf("n1.Consumption = %v, want 24.5", n1.Consumption)
	}
	if n1.Cost == nil || *n1.Cost != 120.4 {
		t.Errorf("n1.Cost = %v, want 120.4", n1.Cost)
	}

	// Check second node (null data)
	n2 := history[1]
	if n2.Consumption != nil {
		t.Errorf("n2.Consumption = %v, want nil", n2.Consumption)
	}
	if n2.Cost != nil {
		t.Errorf("n2.Cost = %v, want nil", n2.Cost)
	}
}

// pagingServer fakes Tibber's cursor pagination over total hourly Nodes, the
// newest ending at end. Cursors are Node indexes. It records the variables of
// every request it serves.
type pagingServer struct {
	*httptest.Server
	start    time.Time
	requests []map[string]interface{}
}

func newPagingServer(t *testing.T, total int, end time.Time) *pagingServer {
	t.Helper()
	ps := &pagingServer{start: end.Add(-time.Duration(total) * time.Hour)}
	ps.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		ps.requests = append(ps.requests, req.Variables)

		last := int(req.Variables["last"].(float64))
		if last > MaxConsumptionPage {
			last = MaxConsumptionPage
		}
		hi := total
		if before, ok := req.Variables["before"].(string); ok {
			hi, _ = strconv.Atoi(before)
		}
		lo := hi - last
		if lo < 0 {
			lo = 0
		}

		nodes := make([]models.ConsumptionNode, 0, hi-lo)
		for i := lo; i < hi; i++ {
			kwh := float64(i)
			nodes = append(nodes, models.ConsumptionNode{
				From:        ps.nodeFrom(i),
				To:          ps.nodeFrom(i + 1),
				Consumption: &kwh,
				Currency:    "NOK",
			})
		}
		history := models.ConsumptionHistory{
			PageInfo: models.PageInfo{StartCursor: strconv.Itoa(lo), HasPreviousPage: lo > 0},
			Nodes:    nodes,
		}
		body, _ := json.Marshal(map[string]interface{}{
			"data": map[string]interface{}{
				"viewer": map[string]interface{}{
					"home": map[string]interface{}{"id": "home-123", "consumption": history},
				},
			},
		})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(ps.Close)
	return ps
}

// nodeFrom is the start of Node i
func (ps *pagingServer) nodeFrom(i int) time.Time {
	return ps.start.Add(time.Duration(i) * time.Hour)
}

func (ps *pagingServer) client() *Client {
	client := NewClient("test-token")
	client.endpoint = ps.URL
	return client
}

// assertContiguous checks that nodes are oldest-first, one hour apart, with no
// gaps or duplicates, starting at first
func assertContiguous(t *testing.T, nodes []models.ConsumptionNode, first time.Time) {
	t.Helper()
	for i, n := range nodes {
		want := first.Add(time.Duration(i) * time.Hour)
		if !n.From.Equal(want) {
			t.Fatalf("nodes[%d].From = %s, want %s", i, n.From, want)
		}
	}
}

func TestClient_GetConsumptionHistory_LastWithinOnePage(t *testing.T) {
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 5000, end)

	nodes, short, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", ConsumptionRange{Last: 744})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if short {
		t.Error("short = true, want false")
	}
	if len(ps.requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(ps.requests))
	}
	if _, ok := ps.requests[0]["before"]; ok {
		t.Error("first request sent a before cursor")
	}
	if len(nodes) != 744 {
		t.Fatalf("len(nodes) = %d, want 744", len(nodes))
	}
	assertContiguous(t, nodes, ps.nodeFrom(5000-744))
}

func TestClient_GetConsumptionHistory_LastPagesPastLimit(t *testing.T) {
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 5000, end)

	nodes, short, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", ConsumptionRange{Last: 2000})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if short {
		t.Error("short = true, want false")
	}
	if len(nodes) != 2000 {
		t.Fatalf("len(nodes) = %d, want 2000", len(nodes))
	}
	assertContiguous(t, nodes, ps.nodeFrom(3000))

	// 744 + 744 + 512: the last page asks only for what is still missing
	wantLast := []float64{744, 744, 512}
	if len(ps.requests) != len(wantLast) {
		t.Fatalf("requests = %d, want %d", len(ps.requests), len(wantLast))
	}
	for i, want := range wantLast {
		if got := ps.requests[i]["last"]; got != want {
			t.Errorf("request %d last = %v, want %v", i, got, want)
		}
	}
}

func TestClient_GetConsumptionHistory_LastLongerThanHistory(t *testing.T) {
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 1000, end)

	nodes, short, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", ConsumptionRange{Last: 2000})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if !short {
		t.Error("short = false, want true")
	}
	if len(nodes) != 1000 {
		t.Fatalf("len(nodes) = %d, want 1000", len(nodes))
	}
	assertContiguous(t, nodes, ps.nodeFrom(0))
}

func TestClient_GetConsumptionHistory_Range(t *testing.T) {
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		from, to  int // Node indexes; to < 0 means no --to
		wantCount int
	}{
		// 5000 - 744*2 = 3512 is a page boundary
		{"from on a page boundary", 3512, -1, 1488},
		{"from just before a page boundary", 3511, -1, 1489},
		{"from just after a page boundary", 3513, -1, 1487},
		{"to on a page boundary", 1000, 3512, 2512},
		{"to inside the newest page", 4000, 4500, 500},
		{"whole history", 0, -1, 5000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := newPagingServer(t, 5000, end)
			r := ConsumptionRange{From: ps.nodeFrom(tt.from)}
			if tt.to >= 0 {
				r.To = ps.nodeFrom(tt.to)
			}

			nodes, short, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", r)
			if err != nil {
				t.Fatalf("GetConsumptionHistory() error = %v", err)
			}
			if short {
				t.Error("short = true, want false")
			}
			if len(nodes) != tt.wantCount {
				t.Fatalf("len(nodes) = %d, want %d", len(nodes), tt.wantCount)
			}
			assertContiguous(t, nodes, r.From)
		})
	}
}

func TestClient_GetConsumptionHistory_RangeStopsPaging(t *testing.T) {
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 50000, end)

	// 1000 Nodes back needs two pages, not the 68 the whole history would take
	_, _, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", ConsumptionRange{From: ps.nodeFrom(49000)})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if len(ps.requests) != 2 {
		t.Errorf("requests = %d, want 2", len(ps.requests))
	}
}

func TestClient_GetConsumptionHistory_RangeBeforeHistory(t *testing.T) {
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 1000, end)

	nodes, short, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY",
		ConsumptionRange{From: ps.nodeFrom(-500)})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if !short {
		t.Error("short = false, want true")
	}
	if len(nodes) != 1000 {
		t.Fatalf("len(nodes) = %d, want 1000", len(nodes))
	}
}

func TestClient_GetConsumptionHistory_OverlappingPages(t *testing.T) {
	// A server whose cursor is one Node too late repeats a Node at every page
	// boundary; the client must not return it twice.
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 2000, end)
	inner := ps.Config.Handler
	ps.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()
		inner.ServeHTTP(rec, r)
		var resp struct {
			Data struct {
				Viewer struct {
					Home struct {
						Consumption models.ConsumptionHistory `json:"consumption"`
					} `json:"home"`
				} `json:"viewer"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		h := resp.Data.Viewer.Home.Consumption
		start, _ := strconv.Atoi(h.PageInfo.StartCursor)
		if h.PageInfo.HasPreviousPage {
			h.PageInfo.StartCursor = strconv.Itoa(start + 1)
		}
		body, _ := json.Marshal(map[string]interface{}{
			"data": map[string]interface{}{"viewer": map[string]interface{}{"home": map[string]interface{}{"consumption": h}}},
		})
		_, _ = w.Write(body)
	})

	nodes, _, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", ConsumptionRange{From: ps.nodeFrom(500)})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if len(nodes) != 1500 {
		t.Fatalf("len(nodes) = %d, want 1500", len(nodes))
	}
	assertContiguous(t, nodes, ps.nodeFrom(500))
}

func TestClient_GetConsumptionHistory_StuckCursor(t *testing.T) {
	// A server that answers every request with the same page and claims there
	// is more must not make the client loop forever.
	end := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	ps := newPagingServer(t, 2000, end)
	inner := ps.Config.Handler
	ps.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = io.NopCloser(strings.NewReader(`{"variables":{"last":744}}`))
		inner.ServeHTTP(w, r)
	})

	nodes, short, err := ps.client().GetConsumptionHistory(context.Background(), "home-123", "HOURLY", ConsumptionRange{Last: 2000})
	if err != nil {
		t.Fatalf("GetConsumptionHistory() error = %v", err)
	}
	if !short {
		t.Error("short = false, want true")
	}
	if len(nodes) != 744 {
		t.Fatalf("len(nodes) = %d, want 744", len(nodes))
	}
	if len(ps.requests) != 2 {
		t.Errorf("requests = %d, want 2", len(ps.requests))
	}
}

func TestClient_GetTimeZone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"data":{"viewer":{"home":{"timeZone":"Europe/Oslo"}}}}`)
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.endpoint = server.URL

	loc, err := client.GetTimeZone(context.Background(), "home-123")
	if err != nil {
		t.Fatalf("GetTimeZone() error = %v", err)
	}
	if loc.String() != "Europe/Oslo" {
		t.Errorf("loc = %q, want Europe/Oslo", loc)
	}
}
