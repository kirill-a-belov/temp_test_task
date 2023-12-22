package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"playGround/modules/client"
	"playGround/utils/tracing"
)

func clientCMD(ctx context.Context) *cobra.Command {
	span, _ := tracing.NewSpan(ctx, "cmd.clientCMD")
	defer span.Close()

	cmd := &cobra.Command{
		Use:   "c",
		Short: "Wisdom client",
		Run: func(cmd *cobra.Command, args []string) {
			logger := log.New(os.Stdout, "client", log.Llongfile)

			config := &client.Config{}
			if err := config.Load(ctx); err != nil {
				logger.Fatal("client config load",
					" error: ", err,
				)
			}

			c := client.New(ctx, config)
			c.Start(ctx)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan,
				syscall.SIGHUP,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)

			select {
			case <-sigChan:
				c.Stop(ctx)
			}
		},
	}

	return cmd
}
