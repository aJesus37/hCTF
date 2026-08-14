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

// --- category ---

var categoryCmd = &cobra.Command{
	Use:     "category",
	Short:   "Manage challenge categories (admin)",
	Aliases: []string{"cat"},
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	RunE:  runCategoryList,
}

var categoryCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a category (admin)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCategoryCreate,
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a category (admin)",
	Args:  cobra.ExactArgs(1),
	RunE:  runCategoryDelete,
}

var categoryUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a category (admin)",
	Args:  cobra.ExactArgs(1),
	RunE:  runCategoryUpdate,
}

var (
	catName  string
	catOrder int
)

// --- difficulty ---

var difficultyCmd = &cobra.Command{
	Use:     "difficulty",
	Short:   "Manage challenge difficulties (admin)",
	Aliases: []string{"diff"},
}

var difficultyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List difficulties",
	RunE:  runDifficultyList,
}

var difficultyCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a difficulty (admin)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDifficultyCreate,
}

var difficultyDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a difficulty (admin)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDifficultyDelete,
}

var difficultyUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a difficulty (admin)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDifficultyUpdate,
}

var (
	diffName      string
	diffOrder     int
	diffColor     string
	diffTextColor string
)

func init() {
	rootCmd.AddCommand(categoryCmd)
	categoryCmd.AddCommand(categoryListCmd, categoryCreateCmd, categoryUpdateCmd, categoryDeleteCmd)
	categoryCreateCmd.Flags().StringVar(&catName, "name", "", "Category name")
	categoryCreateCmd.Flags().IntVar(&catOrder, "order", 0, "Sort order")
	categoryUpdateCmd.Flags().StringVar(&catName, "name", "", "Category name")
	categoryUpdateCmd.Flags().IntVar(&catOrder, "order", 0, "Sort order")

	rootCmd.AddCommand(difficultyCmd)
	difficultyCmd.AddCommand(difficultyListCmd, difficultyCreateCmd, difficultyUpdateCmd, difficultyDeleteCmd)
	difficultyCreateCmd.Flags().StringVar(&diffName, "name", "", "Difficulty name")
	difficultyCreateCmd.Flags().IntVar(&diffOrder, "order", 0, "Sort order")
	difficultyCreateCmd.Flags().StringVar(&diffColor, "color", "", "Badge color (Tailwind classes)")
	difficultyCreateCmd.Flags().StringVar(&diffTextColor, "text-color", "", "Text color (Tailwind class)")
	difficultyUpdateCmd.Flags().StringVar(&diffName, "name", "", "Difficulty name")
	difficultyUpdateCmd.Flags().IntVar(&diffOrder, "order", 0, "Sort order")
	difficultyUpdateCmd.Flags().StringVar(&diffColor, "color", "", "Badge color (Tailwind classes)")
	difficultyUpdateCmd.Flags().StringVar(&diffTextColor, "text-color", "", "Text color (Tailwind class)")
}

// settingItem is a generic representation of a category or difficulty row.
type settingItem struct {
	ID        string
	Name      string
	SortOrder int
}

// printSettingItems renders a table of setting items (categories or difficulties).
func printSettingItems(items []settingItem) {
	cols := []tui.Column{
		{Header: "ID", Width: 34},
		{Header: "NAME", Width: 20},
		{Header: "ORDER", Width: 6},
	}
	var rows [][]string
	for _, item := range items {
		rows = append(rows, []string{tui.Truncate(item.ID, 33), tui.Truncate(item.Name, 19), strconv.Itoa(item.SortOrder)})
	}
	tui.PrintTable(os.Stdout, cols, rows)
}

// runSettingList is a shared list handler for categories and difficulties.
// listFn calls the API and returns raw JSON-encodable items; toItem converts each to settingItem.
func runSettingList[T any](listFn func() ([]T, error), toItem func(T) settingItem) error {
	items, err := listFn()
	if err != nil {
		return err
	}
	if jsonOutput {
		return json.NewEncoder(os.Stdout).Encode(items)
	}
	var rows []settingItem
	for _, item := range items {
		rows = append(rows, toItem(item))
	}
	printSettingItems(rows)
	return nil
}

// runSettingCreate is a shared create handler for categories and difficulties.
// name and order are pointers to the flag variables; label is "Category" or "Difficulty";
// createFn calls the API.
func runSettingCreate(args []string, name *string, order *int, label string, createFn func(*client.Client, string, int) (settingItem, error)) error {
	if len(args) > 0 {
		*name = args[0]
	}
	c, err := newClient()
	if err != nil {
		return err
	}
	if term.IsTerminal(int(os.Stdin.Fd())) && *name == "" {
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title(label + " name").Value(name),
		)).Run(); err != nil {
			return err
		}
	}
	if *name == "" {
		return fmt.Errorf("--name is required")
	}
	item, err := createFn(c, *name, *order)
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, item.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Created %s %q\n", label, item.Name)
	return nil
}

// runSettingDelete is a shared delete handler for categories and difficulties.
func runSettingDelete(args []string, label string, deleteFn func(*client.Client, string) error) error {
	c, err := newClient()
	if err != nil {
		return err
	}
	if err := deleteFn(c, args[0]); err != nil {
		return err
	}
	if !quietOutput {
		fmt.Fprintf(os.Stdout, "Deleted %s %s\n", label, args[0])
	}
	return nil
}

func runCategoryList(_ *cobra.Command, _ []string) error {
	return runSettingList(
		func() ([]client.Category, error) {
			c, err := newClient()
			if err != nil {
				return nil, err
			}
			return c.ListCategories()
		},
		func(cat client.Category) settingItem {
			return settingItem{ID: cat.ID, Name: cat.Name, SortOrder: cat.SortOrder}
		},
	)
}

func runCategoryCreate(_ *cobra.Command, args []string) error {
	return runSettingCreate(args, &catName, &catOrder, "category", func(c *client.Client, name string, order int) (settingItem, error) {
		cat, err := c.CreateCategory(name, order)
		if err != nil {
			return settingItem{}, err
		}
		return settingItem{ID: cat.ID, Name: cat.Name, SortOrder: cat.SortOrder}, nil
	})
}

func runCategoryDelete(_ *cobra.Command, args []string) error {
	return runSettingDelete(args, "category", func(c *client.Client, id string) error {
		return c.DeleteCategory(id)
	})
}

func runDifficultyList(_ *cobra.Command, _ []string) error {
	return runSettingList(
		func() ([]client.Difficulty, error) {
			c, err := newClient()
			if err != nil {
				return nil, err
			}
			return c.ListDifficulties()
		},
		func(d client.Difficulty) settingItem {
			return settingItem{ID: d.ID, Name: d.Name, SortOrder: d.SortOrder}
		},
	)
}

func runDifficultyCreate(_ *cobra.Command, args []string) error {
	return runSettingCreate(args, &diffName, &diffOrder, "difficulty", func(c *client.Client, name string, order int) (settingItem, error) {
		d, err := c.CreateDifficulty(name, order, diffColor, diffTextColor)
		if err != nil {
			return settingItem{}, err
		}
		return settingItem{ID: d.ID, Name: d.Name, SortOrder: d.SortOrder}, nil
	})
}

func runDifficultyDelete(_ *cobra.Command, args []string) error {
	return runSettingDelete(args, "difficulty", func(c *client.Client, id string) error {
		return c.DeleteDifficulty(id)
	})
}

// runCategoryUpdate updates a category. The server's update endpoint is a
// full-replace form handler that requires a non-empty name, so the CLI always
// sends a complete payload: on non-TTY it merges provided flags with the
// current values; on TTY it pre-fills a form and sends all fields.
func runCategoryUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]
	c, err := newClient()
	if err != nil {
		return err
	}
	cats, err := c.ListCategories()
	if err != nil {
		return err
	}
	var cur client.Category
	found := false
	for _, cat := range cats {
		if cat.ID == id {
			cur = cat
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("category %s not found", id)
	}

	// On TTY with no flags, pre-fill a form from the current values.
	if term.IsTerminal(int(os.Stdin.Fd())) && !cmd.Flags().Changed("name") && !cmd.Flags().Changed("order") {
		catName = cur.Name
		catOrder = cur.SortOrder
		orderStr := strconv.Itoa(catOrder)
		if err := huh.NewForm(
			huh.NewGroup(huh.NewInput().Title("Category name").Value(&catName)),
			huh.NewGroup(huh.NewInput().Title("Sort order").Value(&orderStr)),
		).Run(); err != nil {
			return err
		}
		if v, err := strconv.Atoi(orderStr); err == nil {
			catOrder = v
		}
	} else {
		if !cmd.Flags().Changed("name") {
			catName = cur.Name
		}
		if !cmd.Flags().Changed("order") {
			catOrder = cur.SortOrder
		}
	}
	if catName == "" {
		return fmt.Errorf("--name is required")
	}
	cat, err := c.UpdateCategory(id, catName, catOrder)
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, cat.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Updated category %q (%s)\n", cat.Name, cat.ID)
	return nil
}

// runDifficultyUpdate updates a difficulty. Same full-replace semantics as
// runCategoryUpdate: always sends name, color, text_color, and sort_order.
func runDifficultyUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]
	c, err := newClient()
	if err != nil {
		return err
	}
	diffs, err := c.ListDifficulties()
	if err != nil {
		return err
	}
	var cur client.Difficulty
	found := false
	for _, d := range diffs {
		if d.ID == id {
			cur = d
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("difficulty %s not found", id)
	}

	// On TTY with no flags, pre-fill a form from the current values.
	if term.IsTerminal(int(os.Stdin.Fd())) && !cmd.Flags().Changed("name") && !cmd.Flags().Changed("order") && !cmd.Flags().Changed("color") && !cmd.Flags().Changed("text-color") {
		diffName = cur.Name
		diffColor = cur.Color
		diffTextColor = cur.TextColor
		diffOrder = cur.SortOrder
		orderStr := strconv.Itoa(diffOrder)
		if err := huh.NewForm(
			huh.NewGroup(huh.NewInput().Title("Difficulty name").Value(&diffName)),
			huh.NewGroup(huh.NewInput().Title("Badge color (Tailwind classes)").Value(&diffColor)),
			huh.NewGroup(huh.NewInput().Title("Text color (Tailwind class)").Value(&diffTextColor)),
			huh.NewGroup(huh.NewInput().Title("Sort order").Value(&orderStr)),
		).Run(); err != nil {
			return err
		}
		if v, err := strconv.Atoi(orderStr); err == nil {
			diffOrder = v
		}
	} else {
		if !cmd.Flags().Changed("name") {
			diffName = cur.Name
		}
		if !cmd.Flags().Changed("color") {
			diffColor = cur.Color
		}
		if !cmd.Flags().Changed("text-color") {
			diffTextColor = cur.TextColor
		}
		if !cmd.Flags().Changed("order") {
			diffOrder = cur.SortOrder
		}
	}
	if diffName == "" {
		return fmt.Errorf("--name is required")
	}
	d, err := c.UpdateDifficulty(id, diffName, diffColor, diffTextColor, diffOrder)
	if err != nil {
		return err
	}
	if quietOutput {
		fmt.Fprintln(os.Stdout, d.ID)
		return nil
	}
	fmt.Fprintf(os.Stdout, "Updated difficulty %q (%s)\n", d.Name, d.ID)
	return nil
}
