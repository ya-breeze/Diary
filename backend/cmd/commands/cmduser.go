//nolint:forbidigo // it's okay to use fmt in this file
package commands

import (
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
	"github.com/ya-breeze/diary.be/pkg/auth"
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
		Short: "Add a new user",
		RunE: func(_ *cobra.Command, _ []string) error {
			var username string
			fmt.Print("Enter username: ")
			_, err := fmt.Scanln(&username)
			if err != nil {
				return fmt.Errorf("error reading username: %w", err)
			}

			fmt.Print("Enter password: ")
			password, err := gopass.GetPasswd()
			if err != nil {
				return fmt.Errorf("error reading password: %w", err)
			}

			hashed, err := auth.HashPassword(password)
			if err != nil {
				return fmt.Errorf("error hashing password: %w", err)
			}
			fmt.Printf("Use '%s:%s' to add a user\n", username, base64.StdEncoding.EncodeToString(hashed))

			return nil
		},
	}

	return res
}
