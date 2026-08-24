package output

import (
	"fmt"
	"github.com/kristofferrisa/powerctl-cli/internal/models"
	"time"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	FormatHome(home *models.HomeResponse) string
	FormatHomes(homes []models.HomeResponse) string
	FormatPrices(prices *models.PriceInfo, homeID string) string
	FormatLiveMeasurement(m *models.LiveMeasurement) string
	FormatConsumptionHistory(nodes []models.ConsumptionNode, resolution string) string
}

// New creates a formatter based on the format name
func New(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "markdown", "md":
		return &MarkdownFormatter{}
	case "pretty", "":
		return &PrettyFormatter{}
	default:
		return &PrettyFormatter{}
	}
}

// formatPeriod formats the period dynamically based on resolution and in local time
func formatPeriod(from time.Time, to time.Time, resolution string) string {
	fromLocal := from.Local()
	toLocal := to.Local()

	switch resolution {
	case "HOURLY":
		return fromLocal.Format("02 Jan 15:04") + " - " + toLocal.Format("15:04")
	case "DAILY":
		return fromLocal.Format("2006-01-02")
	case "WEEKLY":
		year, week := fromLocal.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case "MONTHLY":
		return fromLocal.Format("2006-01")
	case "ANNUAL":
		return fromLocal.Format("2006")
	default:
		return fromLocal.Format("2006-01-02")
	}
}
