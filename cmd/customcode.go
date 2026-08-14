package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ajesus37/hCTF/internal/tui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var customCodeCmd = &cobra.Command{
	Use:   "custom-code",
	Short: "Manage custom HTML injection settings and MOTD (admin)",
}

var customCodeGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show custom HTML injection settings and MOTD (admin)",
	RunE:  runCustomCodeGet,
}

var customCodeSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Update custom HTML injection settings and MOTD (admin)",
	RunE:  runCustomCodeSet,
}

var (
	ccHeadHTML string
	ccBodyHTML string
	ccPages    string
	ccMotd     string
)

func init() {
	rootCmd.AddCommand(customCodeCmd)
	customCodeCmd.AddCommand(customCodeGetCmd, customCodeSetCmd)
	customCodeSetCmd.Flags().StringVar(&ccHeadHTML, "head-html", "", "HTML to inject in <head>")
	customCodeSetCmd.Flags().StringVar(&ccBodyHTML, "body-end-html", "", "HTML to inject before </body>")
	customCodeSetCmd.Flags().StringVar(&ccPages, "pages", "", "JSON config for per-page injection")
	customCodeSetCmd.Flags().StringVar(&ccMotd, "motd", "", "Message of the day")
}

func runCustomCodeGet(_ *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	cc, err := c.GetCustomCode()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(cc)
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, cc.Motd)
		return nil
	}
	cols := []tui.Column{
		{Header: "FIELD", Width: 20},
		{Header: "VALUE", Width: 60},
	}
	rows := [][]string{
		{"Head HTML", cc.HeadHTML},
		{"Body End HTML", cc.BodyEndHTML},
		{"Pages", cc.PagesJSON},
		{"MOTD", cc.Motd},
	}
	tui.PrintTable(os.Stdout, cols, rows)
	return nil
}

func runCustomCodeSet(cmd *cobra.Command, _ []string) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	tty := term.IsTerminal(int(os.Stdin.Fd()))
	if tty {
		// On TTY, prompt for every field and send all as provided.
		if err := huh.NewForm(
			huh.NewGroup(huh.NewText().Title("Head HTML").Value(&ccHeadHTML)),
			huh.NewGroup(huh.NewText().Title("Body end HTML").Value(&ccBodyHTML)),
			huh.NewGroup(huh.NewInput().Title("Pages JSON").Value(&ccPages)),
			huh.NewGroup(huh.NewInput().Title("MOTD").Value(&ccMotd)),
		).Run(); err != nil {
			return err
		}
	} else if !cmd.Flags().Changed("head-html") && !cmd.Flags().Changed("body-end-html") && !cmd.Flags().Changed("pages") && !cmd.Flags().Changed("motd") {
		return fmt.Errorf("at least one of --head-html, --body-end-html, --pages, --motd is required")
	}
	if err := c.UpdateCustomCode(ccHeadHTML, ccBodyHTML, ccPages, ccMotd,
		tty || cmd.Flags().Changed("head-html"),
		tty || cmd.Flags().Changed("body-end-html"),
		tty || cmd.Flags().Changed("pages"),
		tty || cmd.Flags().Changed("motd"),
	); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintln(os.Stdout, "Custom code updated")
	}
	return nil
}
