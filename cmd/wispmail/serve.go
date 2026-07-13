package commands

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wispmail/wispmail/config"
	"github.com/wispmail/wispmail/internal/app"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Wispmail server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad(configFile)

		ctx, cancel := signal.NotifyContext(
			context.Background(),
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)
		defer cancel()

		application, err := app.New(ctx, cfg)
		if err != nil {
			return err
		}

		log.Printf("Starting Wispmail server (version: %s, commit: %s)", version, commit)

		if err := application.Start(ctx); err != nil {
			return err
		}

		<-ctx.Done()
		log.Println("Shutting down Wispmail server...")

		return application.Stop(context.Background())
	},
}