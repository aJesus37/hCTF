package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ajesus37/hCTF/internal/client"
	"github.com/ajesus37/hCTF/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var competitionCmd = &cobra.Command{Use: "competition", Short: "Competition management", Aliases: []string{"comp"}}
var compListCmd = &cobra.Command{Use: "list", Short: "List competitions", RunE: runCompList}
var compCreateCmd = &cobra.Command{Use: "create [name]", Short: "Create a competition (admin)", Args: cobra.MaximumNArgs(1), RunE: runCompCreate}
var compStartCmd = &cobra.Command{Use: "start <id>", Short: "Force-start a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompStart}
var compEndCmd = &cobra.Command{Use: "end <id>", Short: "Force-end a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompEnd}
var compRegisterCmd = &cobra.Command{Use: "register <id>", Short: "Register your team for a competition", Args: cobra.ExactArgs(1), RunE: runCompRegister}
var compDeleteCmd = &cobra.Command{Use: "delete <id>", Short: "Delete a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompDelete}
var compGetCmd = &cobra.Command{Use: "get <id>", Short: "Get competition details", Args: cobra.ExactArgs(1), RunE: runCompGet}
var compAddChallengeCmd = &cobra.Command{Use: "add-challenge <comp-id> <challenge-id>", Short: "Add challenge to competition (admin)", Args: cobra.ExactArgs(2), RunE: runCompAddChallenge}
var compRemoveChallengeCmd = &cobra.Command{Use: "remove-challenge <comp-id> <challenge-id>", Short: "Remove challenge from competition (admin)", Args: cobra.ExactArgs(2), RunE: runCompRemoveChallenge}
var compFreezeCmd = &cobra.Command{Use: "freeze <id>", Short: "Freeze competition scoreboard (admin)", Args: cobra.ExactArgs(1), RunE: runCompFreeze}
var compUnfreezeCmd = &cobra.Command{Use: "unfreeze <id>", Short: "Unfreeze competition scoreboard (admin)", Args: cobra.ExactArgs(1), RunE: runCompUnfreeze}
var compBlackoutCmd = &cobra.Command{Use: "blackout <id>", Short: "Enable scoreboard blackout for competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompBlackout}
var compUnblackoutCmd = &cobra.Command{Use: "unblackout <id>", Short: "Disable scoreboard blackout for competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompUnblackout}
var compUpdateCmd = &cobra.Command{Use: "update <id>", Short: "Update a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompUpdate}
var compTeamsCmd = &cobra.Command{Use: "teams <id>", Short: "List teams registered for a competition (admin)", Args: cobra.ExactArgs(1), RunE: runCompTeams}
var compScoreboardCmd = &cobra.Command{Use: "scoreboard <id>", Short: "Show scoreboard for a competition", Args: cobra.ExactArgs(1), RunE: runCompScoreboard}

var createCompDescription string
var createCompRulesHTML string
var createCompStart string
var createCompEnd string
var createCompRegStart string
var createCompRegEnd string

var updateCompName string
var updateCompDescription string
var updateCompRulesHTML string
var updateCompStart string
var updateCompEnd string
var updateCompRegStart string
var updateCompRegEnd string

func parseCompetitionID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid competition id: %s", s)
	}
	return id, nil
}

func init() {
	rootCmd.AddCommand(competitionCmd)
	competitionCmd.AddCommand(compListCmd, compCreateCmd, compStartCmd, compEndCmd, compRegisterCmd, compDeleteCmd,
		compGetCmd, compAddChallengeCmd, compRemoveChallengeCmd, compFreezeCmd, compUnfreezeCmd,
		compBlackoutCmd, compUnblackoutCmd, compUpdateCmd, compTeamsCmd, compScoreboardCmd)
	compCreateCmd.Flags().StringVar(&createCompDescription, "description", "", "Competition description")
	compCreateCmd.Flags().StringVar(&createCompRulesHTML, "rules-html", "", "Competition rules (HTML)")
	compCreateCmd.Flags().StringVar(&createCompStart, "start", "", "Start time (RFC3339 or 2006-01-02T15:04)")
	compCreateCmd.Flags().StringVar(&createCompEnd, "end", "", "End time (RFC3339 or 2006-01-02T15:04)")
	compCreateCmd.Flags().StringVar(&createCompRegStart, "registration-start", "", "Registration start time (RFC3339 or 2006-01-02T15:04)")
	compCreateCmd.Flags().StringVar(&createCompRegEnd, "registration-end", "", "Registration end time (RFC3339 or 2006-01-02T15:04)")
	compUpdateCmd.Flags().StringVar(&updateCompName, "name", "", "Competition name")
	compUpdateCmd.Flags().StringVar(&updateCompDescription, "description", "", "Competition description")
	compUpdateCmd.Flags().StringVar(&updateCompRulesHTML, "rules-html", "", "Competition rules (HTML)")
	compUpdateCmd.Flags().StringVar(&updateCompStart, "start", "", "Start time (RFC3339 or 2006-01-02T15:04)")
	compUpdateCmd.Flags().StringVar(&updateCompEnd, "end", "", "End time (RFC3339 or 2006-01-02T15:04)")
	compUpdateCmd.Flags().StringVar(&updateCompRegStart, "registration-start", "", "Registration start time (RFC3339 or 2006-01-02T15:04)")
	compUpdateCmd.Flags().StringVar(&updateCompRegEnd, "registration-end", "", "Registration end time (RFC3339 or 2006-01-02T15:04)")
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
		{Header: "ID", Width: 6},
		{Header: "NAME", Width: 30},
		{Header: "STATUS", Width: 12},
	}
	var rows [][]string
	for _, co := range comps {
		rows = append(rows, []string{strconv.FormatInt(co.ID, 10), co.Name, co.Status})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runCompCreate(_ *cobra.Command, args []string) error {
	createCompName := ""
	if len(args) > 0 {
		createCompName = args[0]
	}

	if term.IsTerminal(int(os.Stdin.Fd())) && createCompDescription == "" {
		groups := []*huh.Group{
			huh.NewGroup(huh.NewText().Title("Description").Value(&createCompDescription)),
			huh.NewGroup(huh.NewText().Title("Rules (HTML)").Value(&createCompRulesHTML)),
			huh.NewGroup(huh.NewInput().Title("Start (RFC3339)").Value(&createCompStart)),
			huh.NewGroup(huh.NewInput().Title("End (RFC3339)").Value(&createCompEnd)),
			huh.NewGroup(huh.NewInput().Title("Registration start (RFC3339)").Value(&createCompRegStart)),
			huh.NewGroup(huh.NewInput().Title("Registration end (RFC3339)").Value(&createCompRegEnd)),
		}
		if createCompName == "" {
			groups = append([]*huh.Group{huh.NewGroup(huh.NewInput().Title("Name").Value(&createCompName))}, groups...)
		}
		if err := huh.NewForm(groups...).Run(); err != nil {
			return err
		}
	}

	if createCompName == "" {
		return fmt.Errorf("name is required")
	}

	c, err := newClient()
	if err != nil {
		return err
	}
	co, err := c.CreateCompetition(createCompName, createCompDescription, createCompRulesHTML, flagIfNonEmpty(createCompStart), flagIfNonEmpty(createCompEnd), flagIfNonEmpty(createCompRegStart), flagIfNonEmpty(createCompRegEnd))
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, co.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Created competition %q (id: %d)\n", co.Name, co.ID)
	return nil
}

func runCompStart(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.ForceStartCompetition(id); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Started competition %d\n", id)
	}
	return nil
}

func runCompEnd(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.ForceEndCompetition(id); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Ended competition %d\n", id)
	}
	return nil
}

func runCompRegister(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.RegisterForCompetition(id); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Registered for competition %d\n", id)
	}
	return nil
}

func runCompDelete(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	ok, err := confirmIfTTY(fmt.Sprintf("Delete competition %d? This cannot be undone.", id))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := c.DeleteCompetition(id); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Deleted competition %d\n", id)
	}
	return nil
}

func runCompGet(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	co, err := c.GetCompetition(id)
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(co)
	}
	cols := []tui.Column{
		{Header: "FIELD", Width: 20},
		{Header: "VALUE", Width: 40},
	}
	rows := [][]string{
		{"ID", strconv.FormatInt(co.ID, 10)},
		{"Name", co.Name},
		{"Status", co.Status},
		{"Scoreboard Frozen", boolToYesNo(co.ScoreboardFrozen)},
	}
	if co.StartAt != "" {
		rows = append(rows, []string{"Start At", co.StartAt})
	}
	if co.EndAt != "" {
		rows = append(rows, []string{"End At", co.EndAt})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runCompAddChallenge(_ *cobra.Command, args []string) error {
	compID, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.AddChallengeToCompetition(compID, args[1]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Added challenge %s to competition %d\n", args[1], compID)
	}
	return nil
}

func runCompRemoveChallenge(_ *cobra.Command, args []string) error {
	compID, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.RemoveChallengeFromCompetition(compID, args[1]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Removed challenge %s from competition %d\n", args[1], compID)
	}
	return nil
}

// runCompToggle is a shared helper for competition boolean-toggle commands
// (freeze/unfreeze, blackout/unblackout). fn is the client method, msg is the
// success message printed when quietOutput is false.
func runCompToggle(args []string, fn func(int64, bool) error, msg string, value bool) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	if err := fn(id, value); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, msg, id)
	}
	return nil
}

func runCompFreeze(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	return runCompToggle(args, c.SetCompetitionFreeze, "Froze scoreboard for competition %d\n", true)
}

func runCompUnfreeze(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	return runCompToggle(args, c.SetCompetitionFreeze, "Unfroze scoreboard for competition %d\n", false)
}

func runCompBlackout(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	return runCompToggle(args, c.SetCompetitionBlackout, "Enabled scoreboard blackout for competition %d\n", true)
}

func runCompUnblackout(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	return runCompToggle(args, c.SetCompetitionBlackout, "Disabled scoreboard blackout for competition %d\n", false)
}

func runCompUpdate(cmd *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}

	// On TTY with no flags, pre-fill the form from current values and send every
	// field (all non-nil) so nothing is lost.
	if term.IsTerminal(int(os.Stdin.Fd())) && !cmd.Flags().Changed("name") {
		co, err := c.GetCompetition(id)
		if err != nil {
			return err
		}
		updateCompName = co.Name
		updateCompDescription = co.Description
		updateCompRulesHTML = co.RulesHTML
		updateCompStart = co.StartAt
		updateCompEnd = co.EndAt
		updateCompRegStart = co.RegistrationStart
		updateCompRegEnd = co.RegistrationEnd

		if err := huh.NewForm(
			huh.NewGroup(huh.NewInput().Title("Name").Value(&updateCompName)),
			huh.NewGroup(huh.NewText().Title("Description").Value(&updateCompDescription)),
			huh.NewGroup(huh.NewText().Title("Rules (HTML)").Value(&updateCompRulesHTML)),
			huh.NewGroup(huh.NewInput().Title("Start (RFC3339)").Value(&updateCompStart)),
			huh.NewGroup(huh.NewInput().Title("End (RFC3339)").Value(&updateCompEnd)),
			huh.NewGroup(huh.NewInput().Title("Registration start (RFC3339)").Value(&updateCompRegStart)),
			huh.NewGroup(huh.NewInput().Title("Registration end (RFC3339)").Value(&updateCompRegEnd)),
		).Run(); err != nil {
			return err
		}

		co, err = c.UpdateCompetition(id, client.CompetitionUpdate{
			Name:              &updateCompName,
			Description:       &updateCompDescription,
			RulesHTML:         &updateCompRulesHTML,
			StartAt:           &updateCompStart,
			EndAt:             &updateCompEnd,
			RegistrationStart: &updateCompRegStart,
			RegistrationEnd:   &updateCompRegEnd,
		})
		if err != nil {
			return err
		}
		if quietOutput {
			fmt.Fprintln(os.Stdout, co.ID)
			return nil
		}
		fmt.Fprintf(os.Stdout, "Updated competition %q (id: %d)\n", co.Name, co.ID)
		return nil
	}

	// Non-interactive: send only explicitly provided flags.
	u := client.CompetitionUpdate{}
	setStrIfChanged(cmd, &u.Name, "name", updateCompName)
	setStrIfChanged(cmd, &u.Description, "description", updateCompDescription)
	setStrIfChanged(cmd, &u.RulesHTML, "rules-html", updateCompRulesHTML)
	setStrIfChanged(cmd, &u.StartAt, "start", updateCompStart)
	setStrIfChanged(cmd, &u.EndAt, "end", updateCompEnd)
	setStrIfChanged(cmd, &u.RegistrationStart, "registration-start", updateCompRegStart)
	setStrIfChanged(cmd, &u.RegistrationEnd, "registration-end", updateCompRegEnd)
	co, err := c.UpdateCompetition(id, u)
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, co.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Updated competition %q (id: %d)\n", co.Name, co.ID)
	return nil
}

func runCompTeams(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	teams, err := c.ListCompetitionTeams(id)
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(teams)
	}
	cols := []tui.Column{
		{Header: "ID", Width: 10},
		{Header: "NAME", Width: 30},
	}
	var rows [][]string
	for _, t := range teams {
		rows = append(rows, []string{tui.Truncate(t.ID, 10), t.Name})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runCompScoreboard(_ *cobra.Command, args []string) error {
	id, err := parseCompetitionID(args[0])
	if err != nil {
		return err
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	entries, err := c.GetCompetitionScoreboard(id)
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
		{Header: "TEAM", Width: 30},
		{Header: "SCORE", Width: 8},
		{Header: "SOLVES", Width: 7},
	}
	var rows [][]string
	for _, e := range entries {
		rows = append(rows, []string{
			strconv.Itoa(e.Rank),
			e.TeamName,
			strconv.Itoa(e.Score),
			strconv.Itoa(e.SolveCount),
		})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}
