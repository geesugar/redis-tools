package main

import (
	check_slots_consistency "github.com/geesugar/redis-tools/check-slots-consistency"
	migrate_slots "github.com/geesugar/redis-tools/migrate-slots"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "redis-tools",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}

	rootCmd.AddCommand(migrate_slots.NewMigrationSlotsCmd())
	rootCmd.AddCommand(check_slots_consistency.NewCheckSlotsConsistencyCmd())

	rootCmd.Execute()
}
