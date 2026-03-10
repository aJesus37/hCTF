package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ajesus37/hCTF2/internal/tui"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{Use: "user", Short: "User management (admin only)"}
var userListCmd = &cobra.Command{Use: "list", Short: "List all users", RunE: runUserList}
var userPromoteCmd = &cobra.Command{Use: "promote <id>", Short: "Grant admin to user", Args: cobra.ExactArgs(1), RunE: runUserPromote}
var userDemoteCmd = &cobra.Command{Use: "demote <id>", Short: "Revoke admin from user", Args: cobra.ExactArgs(1), RunE: runUserDemote}
var userDeleteCmd = &cobra.Command{Use: "delete <id>", Short: "Delete a user", Args: cobra.ExactArgs(1), RunE: runUserDelete}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userListCmd, userPromoteCmd, userDemoteCmd, userDeleteCmd)
}

func runUserList(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	users, err := c.ListUsers()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(users)
	}
	cols := []tui.Column{
		{Header: "ID", Width: 10},
		{Header: "EMAIL", Width: 30},
		{Header: "NAME", Width: 20},
		{Header: "ADMIN", Width: 6},
	}
	var rows [][]string
	for _, u := range users {
		id := u.ID
		if len(id) > 8 {
			id = id[:8] + "..."
		}
		admin := ""
		if u.IsAdmin {
			admin = tui.SolvedStyle.Render("✓")
		}
		rows = append(rows, []string{id, u.Email, u.Name, admin})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runUserPromote(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.PromoteUser(args[0], true); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "User %s promoted to admin\n", args[0])
	}
	return nil
}

func runUserDemote(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.PromoteUser(args[0], false); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "User %s demoted\n", args[0])
	}
	return nil
}

func runUserDelete(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.DeleteUser(args[0]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Deleted user %s\n", args[0])
	}
	return nil
}
