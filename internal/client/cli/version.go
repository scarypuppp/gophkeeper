package cli

import (
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/buildinfo"
	"github.com/spf13/cobra"
)

// VersionCommand печатает сведения о сборке клиента.
func VersionCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Show client build information.",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			table := newTable()
			fmt.Fprintf(table, "Build version:\t%s\n", buildinfo.Version())
			fmt.Fprintf(table, "Build date:\t%s\n", buildinfo.Date())
			fmt.Fprintf(table, "Build commit:\t%s\n", buildinfo.Commit())
			return table.Flush()
		},
	}

	return c
}
