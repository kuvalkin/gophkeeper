package middleware

import "github.com/spf13/cobra"

type CobraRunE func(cmd *cobra.Command, args []string) error

type MW func(CobraRunE) CobraRunE

func Combine(hooks ...CobraRunE) CobraRunE {
	return func(cmd *cobra.Command, args []string) error {
		for _, hook := range hooks {
			if hooks == nil {
				continue
			}

			if err := hook(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}
