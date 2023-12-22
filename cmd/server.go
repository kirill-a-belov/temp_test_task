package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"playGround/modules/server"
	"playGround/utils/tracing"
)

func serverCMD(ctx context.Context) *cobra.Command {
	span, _ := tracing.NewSpan(ctx, "cmd.serverCMD")
	defer span.Close()

	cmd := &cobra.Command{
		Use:   "s",
		Short: "Wisdom server",
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(os.Stdout, "server", log.Llongfile)

			config := &server.Config{}
			if err := config.Load(ctx); err != nil {
				logger.Fatal("server config load",
					" error: ", err,
				)
			}

			s := server.New(ctx, config)
			if err := s.Start(ctx); err != nil {
				logger.Fatal("server staring",
					" error: ", err,
				)
			}

			logger.Println("server has started")

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan,
				syscall.SIGHUP,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)

			select {
			case <-sigChan:
				s.Stop(ctx)
			}
		},
	}

	return cmd
}
