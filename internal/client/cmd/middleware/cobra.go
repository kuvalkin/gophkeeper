package middleware

import "github.com/spf13/cobra"

type CobraRunE func(cmd *cobra.Command, args []string) error

type MW func(CobraRunE) CobraRunE

func Combine(hooks ...CobraRunE) CobraRunE {
	return func(cmd *cobra.Command, args []string) error {
		for _, singleHook := range hooks {
			if singleHook == nil {
				continue
			}

			if err := singleHook(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}
