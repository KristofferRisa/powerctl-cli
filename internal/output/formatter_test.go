package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

func TestNew_ReturnsCorrectFormatter(t *testing.T) {
	tests := []struct {
		format   string
		wantType string
	}{
		{"json", "JSONFormatter"},
		{"markdown", "MarkdownFormatter"},
		{"md", "MarkdownFormatter"},
		{"pretty", "PrettyFormatter"},
		{"", "PrettyFormatter"},
		{"unknown", "PrettyFormatter"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			f := New(tt.format)

			switch tt.wantType {
			case "JSONFormatter":
				if _, ok := f.(*JSONFormatter); !ok {
					t.Errorf("New(%q) = %T, want *JSONFormatter", tt.format, f)
				}
			case "MarkdownFormatter":
				if _, ok := f.(*MarkdownFormatter); !ok {
					t.Errorf("New(%q) = %T, want *MarkdownFormatter", tt.format, f)
				}
			case "PrettyFormatter":
				if _, ok := f.(*PrettyFormatter); !ok {
					t.Errorf("New(%q) = %T, want *PrettyFormatter", tt.format, f)
				}
			}
		})
	}
}

func sampleHome() *models.HomeResponse {
	return &models.HomeResponse{
		Home: models.Home{
			ID:          "home-123",
			AppNickname: "My House",
			Size:        150,
			Type:        "HOUSE",
			Address: models.Address{
				Address1:   "123 Main St",
				PostalCode: "12345",
				City:       "Oslo",
				Country:    "Norway",
			},
			Features: models.Features{
				RealTimeConsumptionEnabled: true,
			},
		},
	}
}

func samplePrices() *models.PriceInfo {
	now := time.Now()
	return &models.PriceInfo{
		Current: &models.Price{
			Total:    0.45,
			Energy:   0.35,
			Tax:      0.10,
			StartsAt: now,
			Level:    "NORMAL",
			Currency: "NOK",
		},
		Today: []models.Price{
			{Total: 0.40, Level: "CHEAP", StartsAt: now.Add(-1 * time.Hour), Currency: "NOK"},
			{Total: 0.45, Level: "NORMAL", StartsAt: now, Currency: "NOK"},
			{Total: 0.60, Level: "EXPENSIVE", StartsAt: now.Add(1 * time.Hour), Currency: "NOK"},
		},
	}
}

func sampleLiveMeasurement() *models.LiveMeasurement {
	return &models.LiveMeasurement{
		Timestamp:              time.Now(),
		Power:                  1234,
		AccumulatedConsumption: 12.5,
		AccumulatedCost:        45.30,
		VoltagePhase1:          230,
		VoltagePhase2:          231,
		VoltagePhase3:          229,
		CurrentL1:              5.2,
		CurrentL2:              3.1,
		CurrentL3:              4.5,
		Currency:               "NOK",
	}
}

// JSON Formatter Tests

func TestJSONFormatter_FormatHome(t *testing.T) {
	f := &JSONFormatter{}
	home := sampleHome()

	output := f.FormatHome(home)

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("FormatHome() output is not valid JSON: %v", err)
	}

	// Check key fields
	if result["id"] != "home-123" {
		t.Errorf("FormatHome() id = %v, want home-123", result["id"])
	}
	if result["appNickname"] != "My House" {
		t.Errorf("FormatHome() appNickname = %v, want My House", result["appNickname"])
	}
}

func TestJSONFormatter_FormatPrices(t *testing.T) {
	f := &JSONFormatter{}
	prices := samplePrices()

	output := f.FormatPrices(prices, "home-123")

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("FormatPrices() output is not valid JSON: %v", err)
	}

	// Check current price exists
	if result["current"] == nil {
		t.Error("FormatPrices() missing current price")
	}
}

func TestJSONFormatter_FormatLiveMeasurement(t *testing.T) {
	f := &JSONFormatter{}
	m := sampleLiveMeasurement()

	output := f.FormatLiveMeasurement(m)

	// Should be valid JSON (compact, single line)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("FormatLiveMeasurement() output is not valid JSON: %v", err)
	}

	// Should not contain newlines (compact format for streaming)
	if strings.Contains(output, "\n") {
		t.Error("FormatLiveMeasurement() should be compact (no newlines)")
	}

	// Check power value
	if result["power"].(float64) != 1234 {
		t.Errorf("FormatLiveMeasurement() power = %v, want 1234", result["power"])
	}
}

// Markdown Formatter Tests

func TestMarkdownFormatter_FormatHome(t *testing.T) {
	f := &MarkdownFormatter{}
	home := sampleHome()

	output := f.FormatHome(home)

	// Should contain key elements
	if !strings.Contains(output, "My House") {
		t.Error("FormatHome() should contain home nickname")
	}
	if !strings.Contains(output, "home-123") {
		t.Error("FormatHome() should contain home ID")
	}
	if !strings.Contains(output, "150") {
		t.Error("FormatHome() should contain size")
	}
	if !strings.Contains(output, "|") {
		t.Error("FormatHome() should contain table formatting")
	}
}

func TestMarkdownFormatter_FormatPrices(t *testing.T) {
	f := &MarkdownFormatter{}
	prices := samplePrices()

	output := f.FormatPrices(prices, "home-123")

	// Should contain headers
	if !strings.Contains(output, "# ") {
		t.Error("FormatPrices() should contain markdown headers")
	}
	// Should contain price value
	if !strings.Contains(output, "0.45") {
		t.Error("FormatPrices() should contain current price")
	}
	// Should contain table
	if !strings.Contains(output, "|") {
		t.Error("FormatPrices() should contain table formatting")
	}
}

func TestMarkdownFormatter_FormatLiveMeasurement(t *testing.T) {
	f := &MarkdownFormatter{}
	m := sampleLiveMeasurement()

	output := f.FormatLiveMeasurement(m)

	// Should contain power value
	if !strings.Contains(output, "1234") {
		t.Error("FormatLiveMeasurement() should contain power value")
	}
	// Should contain table
	if !strings.Contains(output, "|") {
		t.Error("FormatLiveMeasurement() should contain table formatting")
	}
}

// Pretty Formatter Tests

func TestPrettyFormatter_FormatHome(t *testing.T) {
	f := &PrettyFormatter{}
	home := sampleHome()

	output := f.FormatHome(home)

	// Should contain home name
	if !strings.Contains(output, "My House") {
		t.Error("FormatHome() should contain home nickname")
	}
	// Should contain ANSI color codes (escape sequences)
	if !strings.Contains(output, "\033[") {
		t.Error("FormatHome() should contain ANSI color codes")
	}
	// Should contain Pulse status
	if !strings.Contains(output, "Connected") {
		t.Error("FormatHome() should show Pulse as connected")
	}
}

func TestPrettyFormatter_FormatPrices(t *testing.T) {
	f := &PrettyFormatter{}
	prices := samplePrices()

	output := f.FormatPrices(prices, "home-123")

	// Should contain NOW indicator
	if !strings.Contains(output, "NOW") {
		t.Error("FormatPrices() should contain NOW indicator")
	}
	// Should contain color codes
	if !strings.Contains(output, "\033[") {
		t.Error("FormatPrices() should contain ANSI color codes")
	}
	// Should contain price bars
	if !strings.Contains(output, "█") || !strings.Contains(output, "░") {
		t.Error("FormatPrices() should contain price bar visualization")
	}
}

func TestPrettyFormatter_FormatLiveMeasurement(t *testing.T) {
	f := &PrettyFormatter{}
	m := sampleLiveMeasurement()

	output := f.FormatLiveMeasurement(m)

	// Should contain power with unit
	if !strings.Contains(output, "1234") || !strings.Contains(output, "W") {
		t.Error("FormatLiveMeasurement() should contain power with unit")
	}
	// Should contain color codes
	if !strings.Contains(output, "\033[") {
		t.Error("FormatLiveMeasurement() should contain ANSI color codes")
	}
}

func TestPrettyFormatter_PowerColorByUsage(t *testing.T) {
	f := &PrettyFormatter{}

	tests := []struct {
		name  string
		power float64
		color string
	}{
		{"low power", 500, BrightGreen},
		{"medium power", 3000, BrightYellow},
		{"high power", 6000, BrightRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &models.LiveMeasurement{
				Timestamp: time.Now(),
				Power:     tt.power,
				Currency:  "NOK",
			}
			output := f.FormatLiveMeasurement(m)

			if !strings.Contains(output, tt.color) {
				t.Errorf("Power %.0f should use color %q", tt.power, tt.color)
			}
		})
	}
}

func floatPtr(f float64) *float64 { return &f }

func sampleConsumptionNodes() []models.ConsumptionNode {
	t1, _ := time.Parse(time.RFC3339, "2023-10-01T00:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2023-10-02T00:00:00Z")
	t3, _ := time.Parse(time.RFC3339, "2023-10-03T00:00:00Z")

	return []models.ConsumptionNode{
		{
			From:         t1,
			To:           t2,
			Consumption:  floatPtr(24.5),
			Cost:         floatPtr(120.4),
			UnitPrice:    floatPtr(4.9),
			UnitPriceVAT: floatPtr(1.22),
			Currency:     "NOK",
		},
		{
			From:         t2,
			To:           t3,
			Consumption:  floatPtr(15.0),
			Cost:         floatPtr(60.0),
			UnitPrice:    floatPtr(4.0),
			UnitPriceVAT: floatPtr(1.00),
			Currency:     "NOK",
		},
	}
}

func TestJSONFormatter_FormatConsumptionHistory(t *testing.T) {
	f := &JSONFormatter{}
	nodes := sampleConsumptionNodes()

	output := f.FormatConsumptionHistory(nodes, "DAILY")

	var result []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("FormatConsumptionHistory() output is not valid JSON array: %v\nOutput: %s", err, output)
	}

	if len(result) != 2 {
		t.Errorf("FormatConsumptionHistory() len = %d, want 2", len(result))
	}

	if result[0]["cost"].(float64) != 120.4 {
		t.Errorf("FormatConsumptionHistory() cost = %v, want 120.4", result[0]["cost"])
	}

	// Must be multiline (indented)
	if !strings.Contains(output, "\n") {
		t.Error("FormatConsumptionHistory() output should be indented")
	}
}

func TestPrettyFormatter_FormatConsumptionHistory(t *testing.T) {
	f := &PrettyFormatter{}
	nodes := sampleConsumptionNodes()

	output := f.FormatConsumptionHistory(nodes, "DAILY")

	// Check table columns exist
	if !strings.Contains(output, "Period") || !strings.Contains(output, "Consumption") || !strings.Contains(output, "Total Cost") || !strings.Contains(output, "Avg Price") {
		t.Error("FormatConsumptionHistory() missing expected table columns")
	}

	// Check data is present
	if !strings.Contains(output, "24.5") || !strings.Contains(output, "120.4") {
		t.Error("FormatConsumptionHistory() missing expected values")
	}

	// Check for Totals row
	if !strings.Contains(output, "Totals") {
		t.Error("FormatConsumptionHistory() missing Totals row")
	}
	if !strings.Contains(output, "39.5") { // 24.5 + 15.0
		t.Error("FormatConsumptionHistory() missing correct total consumption")
	}
	if !strings.Contains(output, "180.4") { // 120.4 + 60.0
		t.Error("FormatConsumptionHistory() missing correct total cost")
	}
}

func TestMarkdownFormatter_FormatConsumptionHistory(t *testing.T) {
	f := &MarkdownFormatter{}
	nodes := sampleConsumptionNodes()

	output := f.FormatConsumptionHistory(nodes, "DAILY")

	if !strings.Contains(output, "| Period") {
		t.Error("FormatConsumptionHistory() missing markdown table header")
	}

	if !strings.Contains(output, "120.4") {
		t.Error("FormatConsumptionHistory() missing values")
	}

	// Check for Totals row in markdown too
	if !strings.Contains(output, "Totals") || !strings.Contains(output, "39.5") {
		t.Error("FormatConsumptionHistory() missing Totals row in markdown")
	}
}

func TestFormatPeriod_DynamicResolution(t *testing.T) {
	// Tibber returns bucket boundaries carrying the home's own offset, e.g.
	// "2023-10-01T00:00:00.000+02:00". Everything coarser than HOURLY must be
	// formatted in that offset, not the viewer's, or the label names a
	// different day than the kWh it sits next to.
	//
	// time.Local is pinned to a zone far enough west that a regression to
	// .Local() would pull each boundary back into the previous period,
	// regardless of what zone the machine running the tests is in.
	orig := time.Local
	time.Local = time.FixedZone("TEST-11", -11*60*60)
	t.Cleanup(func() { time.Local = orig })

	cest := time.FixedZone("CEST", 2*60*60)
	cet := time.FixedZone("CET", 1*60*60)

	tests := []struct {
		name       string
		from       time.Time
		resolution string
		want       string
	}{
		{"DAILY", time.Date(2023, 10, 1, 0, 0, 0, 0, cest), "DAILY", "2023-10-01"},
		{"MONTHLY", time.Date(2023, 10, 1, 0, 0, 0, 0, cest), "MONTHLY", "2023-10"},
		{"ANNUAL", time.Date(2024, 1, 1, 0, 0, 0, 0, cet), "ANNUAL", "2024"},
		// 2023-10-02 is a Monday: the first day of ISO week 40.
		{"WEEKLY", time.Date(2023, 10, 2, 0, 0, 0, 0, cest), "WEEKLY", "2023-W40"},
		{"unknown resolution falls back to date", time.Date(2023, 10, 1, 0, 0, 0, 0, cest), "DECADE", "2023-10-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatPeriod(tt.from, tt.from.Add(time.Hour), tt.resolution)
			if got != tt.want {
				t.Errorf("formatPeriod(%s) = %q, want %q", tt.resolution, got, tt.want)
			}
		})
	}

	// HOURLY is the deliberate exception: an hourly label is about when the
	// power was drawn, so it renders in the viewer's zone.
	t.Run("HOURLY renders in local time", func(t *testing.T) {
		from := time.Date(2023, 10, 1, 0, 0, 0, 0, cest)
		got := formatPeriod(from, from.Add(time.Hour), "HOURLY")
		want := "30 Sep 11:00 - 12:00"
		if got != want {
			t.Errorf("formatPeriod(HOURLY) = %q, want %q", got, want)
		}
	})
}
