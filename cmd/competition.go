package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ajesus37/hCTF2/internal/tui"
	"github.com/spf13/cobra"
)

var competitionCmd = &cobra.Command{Use: "competition", Short: "Competition management", Aliases: []string{"comp"}}
var compListCmd = &cobra.Command{Use: "list", Short: "List competitions", RunE: runCompList}
var compCreateCmd = &cobra.Command{Use: "create <name>", Short: "Create a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompCreate}
var compStartCmd = &cobra.Command{Use: "start <id>", Short: "Force-start a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompStart}
var compEndCmd = &cobra.Command{Use: "end <id>", Short: "Force-end a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompEnd}

func init() {
	rootCmd.AddCommand(competitionCmd)
	competitionCmd.AddCommand(compListCmd, compCreateCmd, compStartCmd, compEndCmd)
}

func runCompList(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	comps, err := c.ListCompetitions()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(comps)
	}
	cols := []tui.Column{
		{Header: "ID", Width: 10},
		{Header: "NAME", Width: 30},
		{Header: "STATUS", Width: 12},
	}
	var rows [][]string
	for _, co := range comps {
		id := co.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		rows = append(rows, []string{id, co.Name, co.Status})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runCompCreate(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	co, err := c.CreateCompetition(args[0])
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, co.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Created competition %q (%s)\n", co.Name, co.ID)
	return nil
}

func runCompStart(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.ForceStartCompetition(args[0]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Started competition %s\n", args[0])
	}
	return nil
}

func runCompEnd(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.ForceEndCompetition(args[0]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Ended competition %s\n", args[0])
	}
	return nil
}
