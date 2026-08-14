package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var ctftimeCmd = &cobra.Command{
	Use:   "ctftime",
	Short: "Export scoreboard in CTFtime.org JSON format",
	RunE:  runCtftime,
}

var ctftimeOutput string

func init() {
	rootCmd.AddCommand(ctftimeCmd)
	ctftimeCmd.Flags().StringVar(&ctftimeOutput, "output", "", "Write to file instead of stdout")
}

func runCtftime(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	data, err := c.GetCTFtime()
	if err != nil {
		return err
	}
	if ctftimeOutput != "" {
		if err := os.WriteFile(ctftimeOutput, data, 0644); err != nil {
			return err
		}
		if !quietOutput {
			fmt.Fprintf(os.Stderr, "Exported CTFtime standings to %s\n", ctftimeOutput)
		}
		return nil
	}
	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}
	return nil
}
