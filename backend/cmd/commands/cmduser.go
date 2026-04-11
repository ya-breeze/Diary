//nolint:forbidigo // it's okay to use fmt in this file
package commands

import (
	"fmt"
	"log/slog"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	kinauth "github.com/ya-breeze/kin-core/auth"
)

func CmdUser(log *slog.Logger) *cobra.Command {
	res := &cobra.Command{
		Use:   "user",
		Short: "Control users",
		Run: func(_ *cobra.Command, _ []string) {
		},
	}

	res.AddCommand(NewUserAdd(log))

	return res
}

func NewUserAdd(log *slog.Logger) *cobra.Command {
	res := &cobra.Command{
		Use:   "add",
		Short: "Add a new user (outputs a DIARY_SEED_USERS entry)",
		RunE: func(_ *cobra.Command, _ []string) error {
			var familyName string
			fmt.Print("Enter family name: ")
			_, err := fmt.Scanln(&familyName)
			if err != nil {
				return fmt.Errorf("error reading family name: %w", err)
			}

			var username string
			fmt.Print("Enter username: ")
			_, err = fmt.Scanln(&username)
			if err != nil {
				return fmt.Errorf("error reading username: %w", err)
			}

			fmt.Print("Enter password: ")
			password, err := gopass.GetPasswd()
			if err != nil {
				return fmt.Errorf("error reading password: %w", err)
			}

			hash, hashErr := kinauth.HashPassword(string(password))
			if hashErr != nil {
				return fmt.Errorf("error hashing password: %w", hashErr)
			}

			log.Info("User hash generated", "username", username)
			fmt.Printf("Add to DIARY_SEED_USERS: %s:%s:%s\n", familyName, username, hash)

			return nil
		},
	}

	return res
}
