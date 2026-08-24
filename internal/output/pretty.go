package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

// ANSI color codes
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	BrightRed    = "\033[91m"
	BrightGreen  = "\033[92m"
	BrightYellow = "\033[93m"
	BrightBlue   = "\033[94m"
	BrightCyan   = "\033[96m"
)

// PrettyFormatter outputs data with colors and nice formatting
type PrettyFormatter struct{}

// FormatHome formats a single home with colors
func (f *PrettyFormatter) FormatHome(home *models.HomeResponse) string {
	var sb strings.Builder

	title := home.AppNickname
	if title == "" {
		title = home.Address.Address1
	}
	if title == "" {
		title = "Home"
	}

	fmt.Fprintf(&sb, "\n%s%s %s%s\n", Bold, Cyan, title, Reset)
	fmt.Fprintf(&sb, "%s%s%s\n\n", Dim, strings.Repeat("─", len(title)+2), Reset)

	// Address
	if home.Address.Address1 != "" {
		fmt.Fprintf(&sb, "  %s📍 Address%s\n", Bold, Reset)
		fmt.Fprintf(&sb, "     %s\n", home.Address.Address1)
		if home.Address.PostalCode != "" || home.Address.City != "" {
			fmt.Fprintf(&sb, "     %s %s, %s\n", home.Address.PostalCode, home.Address.City, home.Address.Country)
		}
		sb.WriteString("\n")
	}

	// Details
	fmt.Fprintf(&sb, "  %s🏠 Details%s\n", Bold, Reset)
	if home.Size > 0 {
		fmt.Fprintf(&sb, "     Size:      %s%d m²%s\n", BrightCyan, home.Size, Reset)
	}
	if home.Type != "" {
		fmt.Fprintf(&sb, "     Type:      %s\n", home.Type)
	}
	if home.NumberOfResidents > 0 {
		fmt.Fprintf(&sb, "     Residents: %d\n", home.NumberOfResidents)
	}
	if home.MainFuseSize > 0 {
		fmt.Fprintf(&sb, "     Main Fuse: %d A\n", home.MainFuseSize)
	}
	sb.WriteString("\n")

	// Pulse status
	fmt.Fprintf(&sb, "  %s⚡ Pulse%s\n", Bold, Reset)
	if home.Features.RealTimeConsumptionEnabled {
		fmt.Fprintf(&sb, "     Status: %s● Connected%s\n", BrightGreen, Reset)
	} else {
		fmt.Fprintf(&sb, "     Status: %s○ Not connected%s\n", Dim, Reset)
	}

	fmt.Fprintf(&sb, "\n  %sID: %s%s\n", Dim, home.ID, Reset)

	return sb.String()
}

// FormatHomes formats multiple homes
func (f *PrettyFormatter) FormatHomes(homes []models.HomeResponse) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n%s%s⚡ Tibber Homes%s\n", Bold, Cyan, Reset)
	fmt.Fprintf(&sb, "%s%s%s\n", Dim, strings.Repeat("─", 16), Reset)

	for _, home := range homes {
		sb.WriteString(f.FormatHome(&home))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatPrices formats price info with colors
func (f *PrettyFormatter) FormatPrices(prices *models.PriceInfo, homeID string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n%s%s⚡ Electricity Prices%s\n", Bold, Cyan, Reset)
	fmt.Fprintf(&sb, "%s%s%s\n\n", Dim, strings.Repeat("─", 22), Reset)

	// Current price - big and prominent
	if prices.Current != nil {
		fmt.Fprintf(&sb, "  %s%sNOW%s  ", Bold, BrightYellow, Reset)
		fmt.Fprintf(&sb, "%s%s%.2f %s/kWh%s", Bold, priceColor(prices.Current.Level), prices.Current.Total, prices.Current.Currency, Reset)
		fmt.Fprintf(&sb, "  %s\n\n", levelLabel(prices.Current.Level))
	}

	// Today's prices
	if len(prices.Today) > 0 {
		fmt.Fprintf(&sb, "  %s📅 Today%s\n", Bold, Reset)
		sb.WriteString(f.formatPriceList(prices.Today))
		sb.WriteString("\n")
	}

	// Tomorrow's prices
	if len(prices.Tomorrow) > 0 {
		fmt.Fprintf(&sb, "  %s📅 Tomorrow%s\n", Bold, Reset)
		sb.WriteString(f.formatPriceList(prices.Tomorrow))
	} else {
		fmt.Fprintf(&sb, "  %s📅 Tomorrow%s\n", Bold, Reset)
		fmt.Fprintf(&sb, "     %sNot yet available (published ~13:00)%s\n", Dim, Reset)
	}

	return sb.String()
}

func (f *PrettyFormatter) formatPriceList(prices []models.Price) string {
	var sb strings.Builder

	// Find min/max for highlighting
	var minPrice, maxPrice float64 = prices[0].Total, prices[0].Total
	for _, p := range prices {
		if p.Total < minPrice {
			minPrice = p.Total
		}
		if p.Total > maxPrice {
			maxPrice = p.Total
		}
	}

	for _, p := range prices {
		hour := p.StartsAt.Local().Format("15:04")

		// Highlight current hour
		now := time.Now()
		isCurrent := p.StartsAt.Local().Hour() == now.Hour() &&
			p.StartsAt.Local().Day() == now.Day()

		prefix := "  "
		if isCurrent {
			prefix = fmt.Sprintf("%s▶%s ", BrightYellow, Reset)
		}

		// Price bar visualization
		barWidth := 20
		if maxPrice > minPrice {
			barLen := int(float64(barWidth) * (p.Total - minPrice) / (maxPrice - minPrice))
			if barLen < 1 {
				barLen = 1
			}
			bar := strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)
			sb.WriteString(fmt.Sprintf("   %s%s %s%s%.2f%s %s%s%s\n",
				prefix, hour,
				priceColor(p.Level), bar, p.Total, Reset,
				Dim, p.Currency, Reset))
		} else {
			fmt.Fprintf(&sb, "   %s%s %.2f %s\n", prefix, hour, p.Total, p.Currency)
		}
	}

	return sb.String()
}

// FormatLiveMeasurement formats live data with colors
func (f *PrettyFormatter) FormatLiveMeasurement(m *models.LiveMeasurement) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n%s%s⚡ Live Power%s\n", Bold, Cyan, Reset)
	fmt.Fprintf(&sb, "%s%s%s\n\n", Dim, strings.Repeat("─", 14), Reset)

	// Power - big and prominent
	powerColor := BrightGreen
	if m.Power > 5000 {
		powerColor = BrightRed
	} else if m.Power > 2000 {
		powerColor = BrightYellow
	}

	fmt.Fprintf(&sb, "  %s%s%.0f W%s\n\n", Bold, powerColor, m.Power, Reset)

	// Production if any
	if m.PowerProduction > 0 {
		fmt.Fprintf(&sb, "  %s☀️  Production:%s %.0f W\n", Green, Reset, m.PowerProduction)
	}

	// Today's stats
	fmt.Fprintf(&sb, "  %s📊 Today%s\n", Bold, Reset)
	fmt.Fprintf(&sb, "     Consumed: %s%.2f kWh%s\n", BrightCyan, m.AccumulatedConsumption, Reset)
	fmt.Fprintf(&sb, "     Cost:     %s%.2f %s%s\n", BrightYellow, m.AccumulatedCost, m.Currency, Reset)

	// Voltage and current if available
	if m.VoltagePhase1 > 0 {
		fmt.Fprintf(&sb, "\n  %s🔌 Grid%s\n", Bold, Reset)
		sb.WriteString(fmt.Sprintf("     Voltage: %.0f / %.0f / %.0f V\n",
			m.VoltagePhase1, m.VoltagePhase2, m.VoltagePhase3))
		sb.WriteString(fmt.Sprintf("     Current: %.1f / %.1f / %.1f A\n",
			m.CurrentL1, m.CurrentL2, m.CurrentL3))
	}

	// Timestamp
	fmt.Fprintf(&sb, "\n  %s%s%s\n", Dim, m.Timestamp.Local().Format("15:04:05"), Reset)

	return sb.String()
}

// Helper functions

func priceColor(level string) string {
	switch level {
	case "VERY_CHEAP":
		return BrightGreen
	case "CHEAP":
		return Green
	case "NORMAL":
		return Yellow
	case "EXPENSIVE":
		return Red
	case "VERY_EXPENSIVE":
		return BrightRed
	default:
		return Reset
	}
}

func levelLabel(level string) string {
	switch level {
	case "VERY_CHEAP":
		return fmt.Sprintf("%s● Very Cheap%s", BrightGreen, Reset)
	case "CHEAP":
		return fmt.Sprintf("%s● Cheap%s", Green, Reset)
	case "NORMAL":
		return fmt.Sprintf("%s● Normal%s", Yellow, Reset)
	case "EXPENSIVE":
		return fmt.Sprintf("%s● Expensive%s", Red, Reset)
	case "VERY_EXPENSIVE":
		return fmt.Sprintf("%s● Very Expensive%s", BrightRed, Reset)
	default:
		return level
	}
}

func (f *PrettyFormatter) FormatConsumptionHistory(nodes []models.ConsumptionNode, resolution string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "\n%s%s📊 Consumption History%s\n", Bold, Cyan, Reset)
	fmt.Fprintf(&sb, "%s%s%s\n\n", Dim, strings.Repeat("─", 24), Reset)

	rowFmt := "  %-12s %s%-12s%s %-13s %s%-12s%s %-13s %9s\n"

	fmt.Fprintf(&sb, "%s  📅 Period    ⚡ Consumption             💰 Total Cost             📊 Avg Price%s\n", Bold, Reset)
	fmt.Fprintf(&sb, "  %s%s%s\n", Dim, strings.Repeat("─", 78), Reset)

	// Find max consumption and cost for the bar graph scale
	var maxCons float64
	var maxCost float64
	for _, n := range nodes {
		if n.Consumption != nil && *n.Consumption > maxCons {
			maxCons = *n.Consumption
		}
		if n.Cost != nil && *n.Cost > maxCost {
			maxCost = *n.Cost
		}
	}

	var totalConsumption float64
	var totalCost float64
	currency := ""

	for _, n := range nodes {
		period := formatPeriod(n.From, n.To, resolution)

		if currency == "" && n.Currency != "" {
			currency = n.Currency
		}

		consStr := "-"
		consBarStr := strings.Repeat(" ", 12) // Empty space if no data
		
		if n.Consumption != nil {
			consStr = fmt.Sprintf("%.2f kWh", *n.Consumption)
			totalConsumption += *n.Consumption
			
			// Build consumption bar
			barWidth := 12
			barLen := 0
			if maxCons > 0 {
				barLen = int(float64(barWidth) * (*n.Consumption) / maxCons)
			}
			if barLen < 1 && *n.Consumption > 0 {
				barLen = 1 // Show at least a blip if > 0
			}
			if barLen > barWidth {
				barLen = barWidth
			}
			consBarStr = strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)
		}

		costStr := "-"
		costBarStr := strings.Repeat(" ", 12) // Empty space if no data
		if n.Cost != nil {
			costStr = fmt.Sprintf("%.2f", *n.Cost)
			totalCost += *n.Cost

			// Build cost bar
			barWidth := 12
			barLen := 0
			if maxCost > 0 {
				barLen = int(float64(barWidth) * (*n.Cost) / maxCost)
			}
			if barLen < 1 && *n.Cost > 0 {
				barLen = 1 // Show at least a blip if > 0
			}
			if barLen > barWidth {
				barLen = barWidth
			}
			costBarStr = strings.Repeat("█", barLen) + strings.Repeat("░", barWidth-barLen)
		}

		priceStr := "-"
		if n.UnitPrice != nil {
			priceStr = fmt.Sprintf("%.2f", *n.UnitPrice)
			if currency != "" {
				priceStr += " " + currency
			}
		}

		costWithCurrency := "-"
		if costStr != "-" {
			costWithCurrency = costStr + " " + currency
		}

		fmt.Fprintf(&sb, rowFmt, period, BrightCyan, consBarStr, Reset, consStr, BrightYellow, costBarStr, Reset, costWithCurrency, priceStr)
	}

	fmt.Fprintf(&sb, "  %s%s%s\n", Dim, strings.Repeat("─", 78), Reset)
	
	footerFmt := "  %-12s %-26s %-26s %9s\n"
	fmt.Fprintf(&sb, "%s"+footerFmt+"%s\n", Bold, "Totals",
		fmt.Sprintf("%.2f kWh", totalConsumption),
		fmt.Sprintf("%.2f %s", totalCost, currency),
		"", Reset)

	return sb.String()
}
