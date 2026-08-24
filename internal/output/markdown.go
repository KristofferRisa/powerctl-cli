package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

// MarkdownFormatter outputs data as Markdown tables
type MarkdownFormatter struct{}

// FormatHome formats a single home as Markdown
func (f *MarkdownFormatter) FormatHome(home *models.HomeResponse) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## %s\n\n", homeTitle(home))

	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	fmt.Fprintf(&sb, "| ID | `%s` |\n", home.ID)

	if home.Address.Address1 != "" {
		fmt.Fprintf(&sb, "| Address | %s |\n", formatAddress(&home.Address))
	}
	if home.Size > 0 {
		fmt.Fprintf(&sb, "| Size | %d m² |\n", home.Size)
	}
	if home.Type != "" {
		fmt.Fprintf(&sb, "| Type | %s |\n", home.Type)
	}
	if home.NumberOfResidents > 0 {
		fmt.Fprintf(&sb, "| Residents | %d |\n", home.NumberOfResidents)
	}
	if home.MainFuseSize > 0 {
		fmt.Fprintf(&sb, "| Main Fuse | %d A |\n", home.MainFuseSize)
	}

	pulseStatus := "No"
	if home.Features.RealTimeConsumptionEnabled {
		pulseStatus = "Yes"
	}
	fmt.Fprintf(&sb, "| Pulse Enabled | %s |\n", pulseStatus)

	return sb.String()
}

// FormatHomes formats multiple homes as Markdown
func (f *MarkdownFormatter) FormatHomes(homes []models.HomeResponse) string {
	var sb strings.Builder

	sb.WriteString("# Tibber Homes\n\n")

	for i, home := range homes {
		sb.WriteString(f.FormatHome(&home))
		if i < len(homes)-1 {
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String()
}

// FormatPrices formats price info as Markdown
func (f *MarkdownFormatter) FormatPrices(prices *models.PriceInfo, homeID string) string {
	var sb strings.Builder

	sb.WriteString("# Electricity Prices\n\n")

	// Current price
	if prices.Current != nil {
		sb.WriteString("## Current Price\n\n")
		fmt.Fprintf(&sb, "**%.2f %s/kWh** (%s)\n\n",
			prices.Current.Total,
			prices.Current.Currency,
			levelEmoji(prices.Current.Level))
	}

	// Today's prices
	if len(prices.Today) > 0 {
		sb.WriteString("## Today\n\n")
		sb.WriteString(formatPriceTable(prices.Today))
		sb.WriteString("\n")
	}

	// Tomorrow's prices
	if len(prices.Tomorrow) > 0 {
		sb.WriteString("## Tomorrow\n\n")
		sb.WriteString(formatPriceTable(prices.Tomorrow))
	} else {
		sb.WriteString("*Tomorrow's prices not yet available (published around 13:00)*\n")
	}

	return sb.String()
}

// FormatLiveMeasurement formats live data as Markdown
func (f *MarkdownFormatter) FormatLiveMeasurement(m *models.LiveMeasurement) string {
	var sb strings.Builder

	sb.WriteString("## Live Power\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	fmt.Fprintf(&sb, "| Power | %.0f W |\n", m.Power)
	if m.PowerProduction > 0 {
		fmt.Fprintf(&sb, "| Production | %.0f W |\n", m.PowerProduction)
	}
	fmt.Fprintf(&sb, "| Today | %.2f kWh |\n", m.AccumulatedConsumption)
	fmt.Fprintf(&sb, "| Cost | %.2f %s |\n", m.AccumulatedCost, m.Currency)

	if m.VoltagePhase1 > 0 {
		fmt.Fprintf(&sb, "| Voltage | %.1f / %.1f / %.1f V |\n",
			m.VoltagePhase1, m.VoltagePhase2, m.VoltagePhase3)
	}
	if m.CurrentL1 > 0 {
		fmt.Fprintf(&sb, "| Current | %.1f / %.1f / %.1f A |\n",
			m.CurrentL1, m.CurrentL2, m.CurrentL3)
	}

	fmt.Fprintf(&sb, "| Updated | %s |\n", m.Timestamp.Format(time.RFC3339))

	return sb.String()
}

// Helper functions

func homeTitle(home *models.HomeResponse) string {
	if home.AppNickname != "" {
		return home.AppNickname
	}
	if home.Address.Address1 != "" {
		return home.Address.Address1
	}
	return "Home"
}

func formatAddress(addr *models.Address) string {
	parts := []string{}
	if addr.Address1 != "" {
		parts = append(parts, addr.Address1)
	}
	if addr.PostalCode != "" || addr.City != "" {
		parts = append(parts, fmt.Sprintf("%s %s", addr.PostalCode, addr.City))
	}
	return strings.Join(parts, ", ")
}

func formatPriceTable(prices []models.Price) string {
	var sb strings.Builder

	sb.WriteString("| Time | Price | Level |\n")
	sb.WriteString("|------|-------|-------|\n")

	for _, p := range prices {
		hour := p.StartsAt.Local().Format("15:04")
		fmt.Fprintf(&sb, "| %s | %.2f %s | %s |\n",
			hour, p.Total, p.Currency, levelEmoji(p.Level))
	}

	return sb.String()
}

func levelEmoji(level string) string {
	switch level {
	case "VERY_CHEAP":
		return "VERY_CHEAP"
	case "CHEAP":
		return "CHEAP"
	case "NORMAL":
		return "NORMAL"
	case "EXPENSIVE":
		return "EXPENSIVE"
	case "VERY_EXPENSIVE":
		return "VERY_EXPENSIVE"
	default:
		return level
	}
}

func (f *MarkdownFormatter) FormatConsumptionHistory(nodes []models.ConsumptionNode, resolution string) string {
	var sb strings.Builder

	sb.WriteString("# Consumption History\n\n")
	sb.WriteString("| Period | Consumption (kWh) | Total Cost | Avg Price |\n")
	sb.WriteString("|--------|------------------:|-----------:|----------:|\n")

	var totalConsumption float64
	var totalCost float64
	currency := ""

	for _, n := range nodes {
		period := formatPeriod(n.From, n.To, resolution)

		consStr := "-"
		if n.Consumption != nil {
			consStr = fmt.Sprintf("%.2f", *n.Consumption)
			totalConsumption += *n.Consumption
		}

		costStr := "-"
		if n.Cost != nil {
			costStr = fmt.Sprintf("%.2f", *n.Cost)
			totalCost += *n.Cost
			if currency == "" && n.Currency != "" {
				currency = n.Currency
			}
		}

		priceStr := "-"
		if n.UnitPrice != nil {
			priceStr = fmt.Sprintf("%.2f", *n.UnitPrice)
		}

		fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n", period, consStr, costStr, priceStr)
	}

	fmt.Fprintf(&sb, "| **Totals** | **%.2f** | **%.2f %s** | |\n", totalConsumption, totalCost, currency)

	return sb.String()
}
