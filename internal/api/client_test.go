package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		w.Write([]byte(response))
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
		w.Write([]byte(response))
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
		w.Write([]byte("Unauthorized"))
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
		w.Write([]byte(response))
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
		w.Write([]byte(response))
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
