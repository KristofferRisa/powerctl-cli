package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Embeds the time zone database, so --from/--to can be read in the home's
	// time zone on systems without one installed (Windows).
	_ "time/tzdata"

	"github.com/spf13/cobra"

	"github.com/kristofferrisa/powerctl-cli/internal/api"
	"github.com/kristofferrisa/powerctl-cli/internal/models"
)

var (
	resolution string
	last       int
	fromFlag   string
	toFlag     string
	homeIDFlag string
)

// dateLayout is the date-only form accepted by --from and --to
const dateLayout = "2006-01-02"

var consumptionCmd = &cobra.Command{
	Use:   "consumption",
	Short: "Show consumption history",
	Long: `Display historical power consumption and cost.

Bound the Consumption History either with --last, or with --from and an
optional --to. --from and --to take YYYY-MM-DD, read as midnight in the home's
time zone, or an RFC 3339 timestamp. --from is inclusive and --to is exclusive.

Tibber returns at most 744 Nodes per request, so longer ranges are fetched in
several requests.`,
	Example: `  powerctl consumption --resolution hourly --from 2026-01-01
  powerctl consumption --resolution hourly --from 2026-01-01 --to 2026-07-01
  powerctl consumption --resolution hourly --last 2000`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cfg.Validate(); err != nil {
			exitWithError("%v", err)
		}

		client := api.NewClient(cfg.Token)
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Fallback to configured home if flag not set
		targetHome := homeIDFlag
		if targetHome == "" {
			targetHome = cfg.HomeID
		}

		// Normalize resolution to uppercase for GraphQL
		res := strings.ToUpper(resolution)

		// Validate resolution
		switch res {
		case "HOURLY", "DAILY", "WEEKLY", "MONTHLY", "ANNUAL":
			// valid
		default:
			exitWithError("Invalid resolution: %q. Must be one of: hourly, daily, weekly, monthly, annual", resolution)
		}

		r, err := consumptionRange(last, cmd.Flags().Changed("last"), fromFlag, toFlag, func() (*time.Location, error) {
			return client.GetTimeZone(ctx, targetHome)
		})
		if err != nil {
			exitWithError("%v", err)
		}

		history, short, err := client.GetConsumptionHistory(ctx, targetHome, res, r)
		if err != nil {
			exitWithError("Failed to fetch consumption history: %v", err)
		}
		if short {
			fmt.Fprintln(os.Stderr, shortHistoryNotice(history))
		}

		fmt.Println(formatter.FormatConsumptionHistory(history, res))
	},
}

// consumptionRange turns the --last, --from and --to flags into a range to
// fetch. homeLocation is only called when a bound is a date without a time,
// since that is read in the home's time zone.
func consumptionRange(last int, lastSet bool, from, to string, homeLocation func() (*time.Location, error)) (api.ConsumptionRange, error) {
	if to != "" && from == "" {
		return api.ConsumptionRange{}, errors.New("--to requires --from")
	}
	if from == "" {
		if last <= 0 {
			return api.ConsumptionRange{}, fmt.Errorf("invalid --last value %d: must be greater than 0", last)
		}
		return api.ConsumptionRange{Last: last}, nil
	}
	if lastSet {
		return api.ConsumptionRange{}, errors.New("--from and --last cannot be used together; pick one way to bound the range")
	}

	var loc *time.Location
	parse := func(flag, value string) (time.Time, error) {
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			return t, nil
		}
		if _, err := time.Parse(dateLayout, value); err != nil {
			return time.Time{}, fmt.Errorf("invalid %s %q: use YYYY-MM-DD or an RFC 3339 timestamp", flag, value)
		}
		if loc == nil {
			var err error
			if loc, err = homeLocation(); err != nil {
				return time.Time{}, fmt.Errorf("failed to read the home's time zone: %w", err)
			}
		}
		return time.ParseInLocation(dateLayout, value, loc)
	}

	r := api.ConsumptionRange{}
	var err error
	if r.From, err = parse("--from", from); err != nil {
		return api.ConsumptionRange{}, err
	}
	if to != "" {
		if r.To, err = parse("--to", to); err != nil {
			return api.ConsumptionRange{}, err
		}
		if !r.To.After(r.From) {
			return api.ConsumptionRange{}, fmt.Errorf("--to (%s) must be after --from (%s)", to, from)
		}
	}
	return r, nil
}

// shortHistoryNotice explains why fewer Nodes came back than were asked for
func shortHistoryNotice(nodes []models.ConsumptionNode) string {
	if len(nodes) == 0 {
		return "Notice: Tibber has no Consumption History for the requested range"
	}
	return fmt.Sprintf("Notice: Tibber's Consumption History starts at %s, so the requested range is only partly covered (%d Nodes)",
		nodes[0].From.Format(time.RFC3339), len(nodes))
}

func init() {
	consumptionCmd.Flags().StringVar(&resolution, "resolution", "daily", "Energy resolution (hourly, daily, weekly, monthly, annual)")
	consumptionCmd.Flags().IntVar(&last, "last", 7, "Number of previous periods to fetch")
	consumptionCmd.Flags().StringVar(&fromFlag, "from", "", "Start of the range, inclusive (YYYY-MM-DD in the home's time zone, or RFC 3339)")
	consumptionCmd.Flags().StringVar(&toFlag, "to", "", "End of the range, exclusive (default: now; requires --from)")
	consumptionCmd.Flags().StringVar(&homeIDFlag, "home-id", "", "Specific home ID to fetch data for (overrides config)")
	rootCmd.AddCommand(consumptionCmd)
}
