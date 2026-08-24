package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kristofferrisa/powerctl-cli/internal/api"
)

var (
	resolution string
	last       int
	homeIDFlag string
)

var consumptionCmd = &cobra.Command{
	Use:   "consumption",
	Short: "Show consumption history",
	Long:  `Display historical power consumption and cost.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cfg.Validate(); err != nil {
			exitWithError("%v", err)
		}

		client := api.NewClient(cfg.Token)
		ctx := context.Background()

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

		if last <= 0 {
			exitWithError("Invalid last value: %d. Must be greater than 0", last)
		}

		history, err := client.GetConsumptionHistory(ctx, targetHome, res, last)
		if err != nil {
			exitWithError("Failed to fetch consumption history: %v", err)
		}

		if len(history) == 0 {
			exitWithError("No consumption data found for the given period")
		}

		fmt.Println(formatter.FormatConsumptionHistory(history, res))
	},
}

func init() {
	consumptionCmd.Flags().StringVar(&resolution, "resolution", "daily", "Energy resolution (hourly, daily, weekly, monthly, annual)")
	consumptionCmd.Flags().IntVar(&last, "last", 7, "Number of previous periods to fetch")
	consumptionCmd.Flags().StringVar(&homeIDFlag, "home-id", "", "Specific home ID to fetch data for (overrides config)")
	rootCmd.AddCommand(consumptionCmd)
}
