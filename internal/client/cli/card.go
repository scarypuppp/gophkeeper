package cli

import (
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// CardCommand объединяет команды работы с банковскими картами.
// Имя карты является её идентификатором и передаётся первым аргументом.
func CardCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "card",
		Short: "Manage bank cards stored in gophkeeper.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	c.AddCommand(addCardCommand(a))
	c.AddCommand(getCardCommand(a))
	c.AddCommand(listCardsCommand(a))
	c.AddCommand(updateCardCommand(a))
	c.AddCommand(deleteCardCommand(a))

	return c
}

func addCardCommand(a *app.App) *cobra.Command {
	var (
		number   string
		holder   string
		expires  string
		cvv      string
		metadata string
	)

	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Add new bank card.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			card, err := a.Storage.AddCard(storage.Card{
				Name:      args[0],
				Number:    number,
				Holder:    holder,
				ExpiresAt: expires,
				CVV:       cvv,
				Metadata:  metadata,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Successfully added card '%s'\n", card.Name)
			return nil
		},
	}

	c.Flags().StringVarP(&number, "number", "n", "", "Card number")
	c.Flags().StringVar(&holder, "holder", "", "Card holder name")
	c.Flags().StringVarP(&expires, "expires", "e", "", "Expiration date as printed on the card, e.g. 12/29")
	c.Flags().StringVar(&cvv, "cvv", "", "Card verification value")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "Arbitrary metadata")
	_ = c.MarkFlagRequired("number")
	_ = c.MarkFlagRequired("holder")
	_ = c.MarkFlagRequired("expires")
	_ = c.MarkFlagRequired("cvv")

	return c
}

func getCardCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "get <name>",
		Short: "Show bank card by name.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			card, err := a.Storage.GetCard(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("name:    %s\n", card.Name)
			fmt.Printf("number:  %s\n", card.Number)
			fmt.Printf("holder:  %s\n", card.Holder)
			fmt.Printf("expires: %s\n", card.ExpiresAt)
			fmt.Printf("cvv:     %s\n", card.CVV)
			fmt.Printf("meta:    %s\n", card.Metadata)
			return nil
		},
	}

	return c
}

func listCardsCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List stored bank cards.",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			cards := a.Storage.Cards()
			if len(cards) == 0 {
				fmt.Println("No cards stored yet")
				return nil
			}

			table := newTable()
			fmt.Fprintln(table, "name\tnumber\texp\tmeta")
			for _, card := range cards {
				fmt.Fprintf(table, "%s\t%s\t%s\t%s\n",
					card.Name,
					maskCardNumber(card.Number),
					card.ExpiresAt,
					truncate(card.Metadata, metaPreviewLen),
				)
			}
			return table.Flush()
		},
	}

	return c
}

func updateCardCommand(a *app.App) *cobra.Command {
	var (
		number   string
		holder   string
		expires  string
		cvv      string
		metadata string
	)

	c := &cobra.Command{
		Use:   "update <name>",
		Short: "Update stored bank card.",
		Long:  "Обновление частичное: меняются только явно переданные флаги.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			var patch storage.CardPatch
			if cmd.Flags().Changed("number") {
				patch.Number = &number
			}
			if cmd.Flags().Changed("holder") {
				patch.Holder = &holder
			}
			if cmd.Flags().Changed("expires") {
				patch.ExpiresAt = &expires
			}
			if cmd.Flags().Changed("cvv") {
				patch.CVV = &cvv
			}
			if cmd.Flags().Changed("meta") {
				patch.Metadata = &metadata
			}

			card, err := a.Storage.UpdateCard(args[0], patch)
			if err != nil {
				return err
			}
			fmt.Printf("Successfully updated card '%s'\n", card.Name)
			return nil
		},
	}

	c.Flags().StringVarP(&number, "number", "n", "", "New card number")
	c.Flags().StringVar(&holder, "holder", "", "New card holder name")
	c.Flags().StringVarP(&expires, "expires", "e", "", "New expiration date, e.g. 12/29")
	c.Flags().StringVar(&cvv, "cvv", "", "New card verification value")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "New metadata")

	return c
}

func deleteCardCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete bank card by name.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}
			name := args[0]

			if err := a.Storage.DeleteCard(name); err != nil {
				return err
			}
			fmt.Printf("Successfully deleted card '%s'\n", name)
			return nil
		},
	}

	return c
}
