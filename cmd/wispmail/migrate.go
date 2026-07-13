package commands

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wispmail/wispmail/config"
	"github.com/wispmail/wispmail/internal/storage/postgres"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad(configFile)

		log.Println("Running database migrations...")

		migrator := postgres.NewMigrator(cfg.Database.URL)
		if err := migrator.Up(); err != nil {
			return err
		}

		log.Println("Migrations completed successfully")
		return nil
	},
}