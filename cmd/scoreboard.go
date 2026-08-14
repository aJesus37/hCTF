package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ajesus37/hCTF/internal/tui"
	"github.com/spf13/cobra"
)

var scoreboardCmd = &cobra.Command{
	Use:   "scoreboard",
	Short: "Show the current scoreboard",
	RunE:  runScoreboard,
}
var scoreboardFreezeCmd = &cobra.Command{Use: "freeze", Short: "Freeze the global scoreboard (admin)", RunE: runScoreboardFreeze}
var scoreboardUnfreezeCmd = &cobra.Command{Use: "unfreeze", Short: "Unfreeze the global scoreboard (admin)", RunE: runScoreboardUnfreeze}
var scoreboardAdminVisibilityCmd = &cobra.Command{
	Use:   "admin-visibility <on|off>",
	Short: "Toggle admin visibility on the scoreboard (admin)",
	Args:  cobra.ExactArgs(1),
	RunE:  runScoreboardAdminVisibility,
}
var scoreboardTeamsCmd = &cobra.Command{Use: "teams", Short: "Show the team scoreboard", RunE: runScoreboardTeams}
var scoreboardEvolutionCmd = &cobra.Command{Use: "evolution", Short: "Show score evolution over time", RunE: runScoreboardEvolution}
var scoreboardForceRecordCmd = &cobra.Command{Use: "force-record", Short: "Manually trigger score recording (admin)", RunE: runScoreboardForceRecord}

var evolutionMode string
var evolutionLimit int

func init() {
	rootCmd.AddCommand(scoreboardCmd)
	scoreboardCmd.AddCommand(scoreboardFreezeCmd, scoreboardUnfreezeCmd, scoreboardAdminVisibilityCmd, scoreboardTeamsCmd, scoreboardEvolutionCmd, scoreboardForceRecordCmd)
	scoreboardEvolutionCmd.Flags().StringVar(&evolutionMode, "mode", "individual", "Score mode: individual or team")
	scoreboardEvolutionCmd.Flags().IntVar(&evolutionLimit, "limit", 20, "Number of top entries to include")
}

func runScoreboard(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	entries, err := c.GetScoreboard()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(entries)
	}
	if len(entries) == 0 {
		fmt.Println("No scoreboard entries found.")
		return nil
	}
	cols := []tui.Column{
		{Header: "RANK", Width: 6},
		{Header: "USER", Width: 25},
		{Header: "TEAM", Width: 20},
		{Header: "SCORE", Width: 8},
		{Header: "SOLVES", Width: 7},
	}
	var rows [][]string
	for _, e := range entries {
		team := ""
		if e.TeamName != nil {
			team = *e.TeamName
		}
		rows = append(rows, []string{
			strconv.Itoa(e.Rank),
			e.UserName,
			team,
			strconv.Itoa(e.Points),
			strconv.Itoa(e.SolveCount),
		})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runScoreboardFreeze(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.SetScoreboardFreeze(true); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintln(os.Stdout, "Scoreboard frozen")
	}
	return nil
}

func runScoreboardUnfreeze(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.SetScoreboardFreeze(false); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintln(os.Stdout, "Scoreboard unfrozen")
	}
	return nil
}

func runScoreboardAdminVisibility(_ *cobra.Command, args []string) error {
	arg := strings.ToLower(args[0])
	if arg != "on" && arg != "off" {
		return fmt.Errorf("expected 'on' or 'off', got %q", args[0])
	}
	visible := arg == "on"
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.SetAdminVisibility(visible); err != nil {
		return err
	}
	if !quietOutput {
		if visible {
			fmt.Fprintln(os.Stdout, "Admins visible on scoreboard")
		} else {
			fmt.Fprintln(os.Stdout, "Admins hidden on scoreboard")
		}
	}
	return nil
}

func runScoreboardTeams(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	entries, err := c.GetTeamScoreboard()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(entries)
	}
	if len(entries) == 0 {
		fmt.Println("No team scoreboard entries found.")
		return nil
	}
	cols := []tui.Column{
		{Header: "RANK", Width: 6},
		{Header: "TEAM", Width: 25},
		{Header: "POINTS", Width: 8},
		{Header: "SOLVES", Width: 7},
	}
	var rows [][]string
	for _, e := range entries {
		team := e.UserName
		if e.TeamName != nil {
			team = *e.TeamName
		}
		rows = append(rows, []string{
			strconv.Itoa(e.Rank),
			tui.Truncate(team, 24),
			strconv.Itoa(e.Points),
			strconv.Itoa(e.SolveCount),
		})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runScoreboardEvolution(_ *cobra.Command, _ []string) error {
	if evolutionMode != "individual" && evolutionMode != "team" {
		return fmt.Errorf("expected --mode 'individual' or 'team', got %q", evolutionMode)
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	resp, err := c.GetScoreEvolution(evolutionMode, evolutionLimit)
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(resp)
	}
	intervals, _ := resp["intervals"].([]interface{})
	series, _ := resp["series"].([]interface{})
	fmt.Fprintf(os.Stdout, "Score evolution (%d intervals, %s)\n", len(intervals), evolutionMode)
	for _, s := range series {
		sm, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := sm["name"].(string)
		scores, _ := sm["scores"].([]interface{})
		last := 0
		if len(scores) > 0 {
			switch n := scores[len(scores)-1].(type) {
			case float64:
				last = int(n)
			case int:
				last = n
			}
		}
		fmt.Fprintf(os.Stdout, "  %s: %d pts\n", name, last)
	}
	return nil
}

func runScoreboardForceRecord(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.ForceScoreRecord(); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintln(os.Stdout, "Score recording triggered")
	}
	return nil
}
