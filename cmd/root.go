package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"playGround/utils/tracing"
)

func New(ctx context.Context) *cobra.Command {
	span, _ := tracing.NewSpan(ctx, "cmd.New")
	defer span.Close()

	cmd := &cobra.Command{
		Short: "Wisdom service command line interface",
	}
	cmd.AddCommand(clientCMD(ctx), serverCMD(ctx))

	return cmd
}
