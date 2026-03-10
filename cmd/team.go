package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ajesus37/hCTF2/internal/tui"
	"github.com/spf13/cobra"
)

var teamCmd = &cobra.Command{Use: "team", Short: "Team management"}
var teamListCmd = &cobra.Command{Use: "list", Short: "List all teams", RunE: runTeamList}
var teamGetCmd = &cobra.Command{Use: "get <id>", Short: "Show team details", Args: cobra.ExactArgs(1), RunE: runTeamGet}
var teamCreateCmd = &cobra.Command{Use: "create <name>", Short: "Create a team", Args: cobra.ExactArgs(1), RunE: runTeamCreate}
var teamJoinCmd = &cobra.Command{Use: "join <invite-code>", Short: "Join a team by invite code", Args: cobra.ExactArgs(1), RunE: runTeamJoin}

func init() {
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamListCmd, teamGetCmd, teamCreateCmd, teamJoinCmd)
}

func runTeamList(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	teams, err := c.ListTeams()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(teams)
	}
	cols := []tui.Column{
		{Header: "ID", Width: 10},
		{Header: "NAME", Width: 25},
		{Header: "SCORE", Width: 8},
		{Header: "MEMBERS", Width: 9},
	}
	var rows [][]string
	for _, t := range teams {
		id := t.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		rows = append(rows, []string{id, t.Name, strconv.Itoa(t.Score), strconv.Itoa(t.MemberCount)})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runTeamGet(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	t, err := c.GetTeam(args[0])
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(t)
	}
	fmt.Fprintf(os.Stdout, "Name:    %s\nID:      %s\nScore:   %d\nMembers: %d\n",
		t.Name, t.ID, t.Score, t.MemberCount)
	if t.InviteCode != "" {
		fmt.Fprintf(os.Stdout, "Invite:  %s\n", t.InviteCode)
	}
	return nil
}

func runTeamCreate(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	t, err := c.CreateTeam(args[0])
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, t.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Created team %q (%s)\n", t.Name, t.ID)
	return nil
}

func runTeamJoin(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.JoinTeam(args[0]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintln(os.Stdout, "Joined team")
	}
	return nil
}
