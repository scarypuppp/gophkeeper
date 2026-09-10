package cli

import (
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// TextCommand объединяет команды работы с произвольными текстовыми данными.
// Имя текста является его идентификатором и передаётся первым аргументом.
func TextCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "text",
		Short: "Manage arbitrary text data stored in gophkeeper.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	c.AddCommand(addTextCommand(a))
	c.AddCommand(getTextCommand(a))
	c.AddCommand(listTextsCommand(a))
	c.AddCommand(updateTextCommand(a))
	c.AddCommand(deleteTextCommand(a))

	return c
}

func addTextCommand(a *app.App) *cobra.Command {
	var (
		text     string
		metadata string
	)

	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Add new text data.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			added, err := a.Storage.AddText(storage.Text{
				Name:     args[0],
				Text:     text,
				Metadata: metadata,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Successfully added text '%s'\n", added.Name)
			return nil
		},
	}

	c.Flags().StringVarP(&text, "text", "t", "", "Text to store")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "Arbitrary metadata")
	_ = c.MarkFlagRequired("text")

	return c
}

func getTextCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "get <name>",
		Short: "Show text data by name.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			text, err := a.Storage.GetText(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("name: %s\n", text.Name)
			fmt.Printf("text: %s\n", text.Text)
			fmt.Printf("meta: %s\n", text.Metadata)
			return nil
		},
	}

	return c
}

func listTextsCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List stored text data.",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			texts := a.Storage.Texts()
			if len(texts) == 0 {
				fmt.Println("No text data stored yet")
				return nil
			}

			table := newTable()
			fmt.Fprintln(table, "name\ttext\tmeta")
			for _, text := range texts {
				fmt.Fprintf(table, "%s\t%s\t%s\n",
					text.Name,
					truncate(text.Text, metaPreviewLen),
					truncate(text.Metadata, metaPreviewLen),
				)
			}
			return table.Flush()
		},
	}

	return c
}

func updateTextCommand(a *app.App) *cobra.Command {
	var (
		text     string
		metadata string
	)

	c := &cobra.Command{
		Use:   "update <name>",
		Short: "Update stored text data.",
		Long:  "Обновление частичное: меняются только явно переданные флаги.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			var patch storage.TextPatch
			if cmd.Flags().Changed("text") {
				patch.Text = &text
			}
			if cmd.Flags().Changed("meta") {
				patch.Metadata = &metadata
			}

			updated, err := a.Storage.UpdateText(args[0], patch)
			if err != nil {
				return err
			}
			fmt.Printf("Successfully updated text '%s'\n", updated.Name)
			return nil
		},
	}

	c.Flags().StringVarP(&text, "text", "t", "", "New text")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "New metadata")

	return c
}

func deleteTextCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete text data by name.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}
			name := args[0]

			if err := a.Storage.DeleteText(name); err != nil {
				return err
			}
			fmt.Printf("Successfully deleted text '%s'\n", name)
			return nil
		},
	}

	return c
}
