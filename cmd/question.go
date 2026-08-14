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

var questionCmd = &cobra.Command{Use: "question", Short: "Manage challenge questions (admin)", Aliases: []string{"q"}}
var questionGetCmd = &cobra.Command{Use: "get <id>", Short: "Show question details (admin)", Args: cobra.ExactArgs(1), RunE: runQuestionGet}
var questionListCmd = &cobra.Command{Use: "list <challenge-id>", Short: "List questions for a challenge", Args: cobra.ExactArgs(1), RunE: runQuestionList}
var questionCreateCmd = &cobra.Command{Use: "create", Short: "Create a question (admin)", RunE: runQuestionCreate}
var questionDeleteCmd = &cobra.Command{Use: "delete <id>", Short: "Delete a question (admin)", Args: cobra.ExactArgs(1), RunE: runQuestionDelete}
var questionUpdateCmd = &cobra.Command{Use: "update <id>", Short: "Update a question (admin)", Args: cobra.ExactArgs(1), RunE: runQuestionUpdate}

var (
	qChallengeID    string
	qName           string
	qDescription    string
	qFlag           string
	qFlagMask       string
	qCaseSensitive  bool
	qPoints         int
	quName          string
	quDescription   string
	quFlag          string
	quFlagMask      string
	quPoints        int
	quCaseSensitive bool
)

func init() {
	rootCmd.AddCommand(questionCmd)
	questionCmd.AddCommand(questionGetCmd, questionListCmd, questionCreateCmd, questionDeleteCmd, questionUpdateCmd)
	questionCreateCmd.Flags().StringVar(&qChallengeID, "challenge", "", "Challenge ID")
	questionCreateCmd.Flags().StringVar(&qName, "name", "", "Question name")
	questionCreateCmd.Flags().StringVar(&qDescription, "description", "", "Question description")
	questionCreateCmd.Flags().StringVar(&qFlag, "flag", "", "Flag value")
	questionCreateCmd.Flags().StringVar(&qFlagMask, "flag-mask", "", "Flag display mask")
	questionCreateCmd.Flags().BoolVar(&qCaseSensitive, "case-sensitive", false, "Case-sensitive flag matching")
	questionCreateCmd.Flags().IntVar(&qPoints, "points", 100, "Point value")

	questionUpdateCmd.Flags().StringVar(&quName, "name", "", "Question name")
	questionUpdateCmd.Flags().StringVar(&quDescription, "description", "", "Question description")
	questionUpdateCmd.Flags().StringVar(&quFlag, "flag", "", "Flag value")
	questionUpdateCmd.Flags().StringVar(&quFlagMask, "flag-mask", "", "Flag display mask")
	questionUpdateCmd.Flags().IntVar(&quPoints, "points", 0, "Point value")
	questionUpdateCmd.Flags().BoolVar(&quCaseSensitive, "case-sensitive", false, "Case-sensitive flag matching")
}

func runQuestionGet(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	q, err := c.GetQuestion(args[0])
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(q)
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, q.ID)
		return nil
	}
	flagMask := ""
	if q.FlagMask != nil {
		flagMask = *q.FlagMask
	}
	cols := []tui.Column{
		{Header: "FIELD", Width: 20},
		{Header: "VALUE", Width: 40},
	}
	rows := [][]string{
		{"ID", q.ID},
		{"Challenge", q.ChallengeID},
		{"Name", q.Name},
		{"Points", strconv.Itoa(q.Points)},
		{"Flag mask", flagMask},
		{"Case sensitive", boolToYesNo(q.CaseSensitive)},
	}
	if q.Description != "" {
		rows = append(rows, []string{"Description", q.Description})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runQuestionList(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	qs, err := c.ListQuestions(args[0])
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(qs)
	}
	cols := []tui.Column{
		{Header: "ID", Width: 10},
		{Header: "NAME", Width: 30},
		{Header: "FLAG MASK", Width: 20},
		{Header: "PTS", Width: 6},
	}
	var rows [][]string
	for _, q := range qs {
		id := tui.Truncate(q.ID, 10)
		mask := ""
		if q.FlagMask != "" {
			mask = q.FlagMask
		}
		rows = append(rows, []string{id, q.Name, mask, strconv.Itoa(q.Points)})
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runQuestionCreate(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if term.IsTerminal(int(os.Stdin.Fd())) && qName == "" {
		pointsStr := strconv.Itoa(qPoints)
		if err := huh.NewForm(
			huh.NewGroup(huh.NewInput().Title("Challenge ID").Value(&qChallengeID)),
			huh.NewGroup(huh.NewInput().Title("Question name").Value(&qName)),
			huh.NewGroup(huh.NewText().Title("Description").Value(&qDescription)),
			huh.NewGroup(huh.NewInput().Title("Flag").Value(&qFlag)),
			huh.NewGroup(huh.NewInput().Title("Flag mask").Value(&qFlagMask)),
			huh.NewGroup(huh.NewInput().Title("Points").Value(&pointsStr)),
			huh.NewGroup(huh.NewConfirm().Title("Case-sensitive flag?").Value(&qCaseSensitive)),
		).Run(); err != nil {
			return err
		}
		if p, err := strconv.Atoi(pointsStr); err == nil {
			qPoints = p
		}
	}
	if qChallengeID == "" || qFlag == "" {
		return fmt.Errorf("--challenge and --flag are required")
	}
	q, err := c.CreateQuestion(qChallengeID, qName, qDescription, qFlag, flagIfNonEmpty(qFlagMask), qCaseSensitive, qPoints)
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, q.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Created question %s (%s)\n", q.Name, q.ID)
	return nil
}

func runQuestionDelete(_ *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := c.DeleteQuestion(args[0]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Deleted question %s\n", args[0])
	}
	return nil
}

func buildQuestionGroups(name, description, flag, flagMask, pointsStr *string, caseSensitive *bool) []*huh.Group {
	return []*huh.Group{
		huh.NewGroup(huh.NewInput().Title("Name").Value(name)),
		huh.NewGroup(huh.NewText().Title("Description").Value(description)),
		huh.NewGroup(huh.NewInput().Title("Flag").Value(flag)),
		huh.NewGroup(huh.NewInput().Title("Flag mask").Value(flagMask)),
		huh.NewGroup(huh.NewInput().Title("Points").Value(pointsStr)),
		huh.NewGroup(huh.NewConfirm().Title("Case-sensitive flag?").Value(caseSensitive)),
	}
}

func runQuestionUpdate(cmd *cobra.Command, args []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	id := args[0]

	if term.IsTerminal(int(os.Stdin.Fd())) && !cmd.Flags().Changed("name") {
		q, err := c.GetQuestion(id)
		if err != nil {
			return err
		}
		quName = q.Name
		quDescription = q.Description
		quCaseSensitive = q.CaseSensitive
		if q.FlagMask != nil {
			quFlagMask = *q.FlagMask
		}
		if quPoints == 0 {
			quPoints = q.Points
		}
		pointsStr := strconv.Itoa(quPoints)
		groups := buildQuestionGroups(&quName, &quDescription, &quFlag, &quFlagMask, &pointsStr, &quCaseSensitive)
		if err := huh.NewForm(groups...).Run(); err != nil {
			return err
		}
		if p, err := strconv.Atoi(pointsStr); err == nil {
			quPoints = p
		}
		// On TTY everything is considered provided.
		u := client.QuestionUpdate{
			Name:          &quName,
			Description:   &quDescription,
			Flag:          flagIfNonEmpty(quFlag),
			FlagMask:      &quFlagMask,
			Points:        &quPoints,
			CaseSensitive: &quCaseSensitive,
		}
		if err := c.UpdateQuestion(id, u); err != nil {
			return err
		}
		if quietOutput {
			fmt.Fprintln(os.Stdout, id)
		} else {
			fmt.Fprintf(os.Stdout, "Updated question %s\n", id)
		}
		return nil
	}

	u := client.QuestionUpdate{}
	setStrIfChanged(cmd, &u.Name, "name", quName)
	setStrIfChanged(cmd, &u.Description, "description", quDescription)
	setStrIfChanged(cmd, &u.Flag, "flag", quFlag)
	setStrIfChanged(cmd, &u.FlagMask, "flag-mask", quFlagMask)
	setIntIfChanged(cmd, &u.Points, "points", quPoints)
	setBoolIfChanged(cmd, &u.CaseSensitive, "case-sensitive", quCaseSensitive)
	if err := c.UpdateQuestion(id, u); err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, id)
	} else {
		fmt.Fprintf(os.Stdout, "Updated question %s\n", id)
	}
	return nil
}
