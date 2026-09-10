package cli

import (
	"fmt"

	"github.com/scarypuppp/gophkeeper/internal/client/app"
	"github.com/scarypuppp/gophkeeper/internal/client/storage"
	"github.com/spf13/cobra"
)

// SecretCommand объединяет команды работы с парами логин-пароль.
// Имя секрета является его идентификатором и передаётся первым аргументом.
func SecretCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "secret",
		Short: "Manage login-password pairs stored in gophkeeper.",
		Long:  "",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	c.AddCommand(addSecretCommand(a))
	c.AddCommand(getSecretCommand(a))
	c.AddCommand(listSecretsCommand(a))
	c.AddCommand(updateSecretCommand(a))
	c.AddCommand(deleteSecretCommand(a))

	return c
}

func addSecretCommand(a *app.App) *cobra.Command {
	var (
		login    string
		password string
		metadata string
	)

	c := &cobra.Command{
		Use:   "add <name>",
		Short: "Add new secret.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			secret, err := a.Storage.AddSecret(storage.Secret{
				Name:     args[0],
				Login:    login,
				Password: password,
				Metadata: metadata,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Successfully added secret '%s'\n", secret.Name)
			return nil
		},
	}

	c.Flags().StringVarP(&login, "login", "l", "", "Login stored in the secret")
	c.Flags().StringVarP(&password, "password", "p", "", "Password stored in the secret")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "Arbitrary metadata")
	_ = c.MarkFlagRequired("login")
	_ = c.MarkFlagRequired("password")

	return c
}

func getSecretCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "get <name>",
		Short: "Show secret by name.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			secret, err := a.Storage.GetSecret(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("name:     %s\n", secret.Name)
			fmt.Printf("login:    %s\n", secret.Login)
			fmt.Printf("password: %s\n", secret.Password)
			fmt.Printf("meta:     %s\n", secret.Metadata)
			return nil
		},
	}

	return c
}

func listSecretsCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List stored secrets.",
		Long:  "",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			secrets := a.Storage.Secrets()
			if len(secrets) == 0 {
				fmt.Println("No secrets stored yet")
				return nil
			}

			table := newTable()
			fmt.Fprintln(table, "name\tlogin\tmeta")
			for _, secret := range secrets {
				fmt.Fprintf(table, "%s\t%s\t%s\n", secret.Name, secret.Login, truncate(secret.Metadata, metaPreviewLen))
			}
			return table.Flush()
		},
	}

	return c
}

func updateSecretCommand(a *app.App) *cobra.Command {
	var (
		login    string
		password string
		metadata string
	)

	c := &cobra.Command{
		Use:   "update <name>",
		Short: "Update stored secret.",
		Long:  "Обновление частичное: меняются только явно переданные флаги.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}

			var patch storage.SecretPatch
			if cmd.Flags().Changed("login") {
				patch.Login = &login
			}
			if cmd.Flags().Changed("password") {
				patch.Password = &password
			}
			if cmd.Flags().Changed("meta") {
				patch.Metadata = &metadata
			}

			secret, err := a.Storage.UpdateSecret(args[0], patch)
			if err != nil {
				return err
			}
			fmt.Printf("Successfully updated secret '%s'\n", secret.Name)
			return nil
		},
	}

	c.Flags().StringVarP(&login, "login", "l", "", "New login")
	c.Flags().StringVarP(&password, "password", "p", "", "New password")
	c.Flags().StringVarP(&metadata, "meta", "m", "", "New metadata")

	return c
}

func deleteSecretCommand(a *app.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete secret by name.",
		Long:  "",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireAuth(a); err != nil {
				return err
			}
			name := args[0]

			if err := a.Storage.DeleteSecret(name); err != nil {
				return err
			}
			fmt.Printf("Successfully deleted secret '%s'\n", name)
			return nil
		},
	}

	return c
}
