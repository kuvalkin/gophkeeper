package middleware

import "github.com/spf13/cobra"

// CobraRunE defines a function type that matches the signature of cobra.Command's RunE field.
// It takes a cobra.Command and its arguments and returns an error.
type CobraRunE func(cmd *cobra.Command, args []string) error

// MW defines a middleware function type that wraps a CobraRunE function.
// It allows chaining or modifying the behavior of CobraRunE functions.
type MW func(CobraRunE) CobraRunE

// Combine combines multiple CobraRunE functions into a single CobraRunE function.
// Each function is executed in the order provided, and if any function returns an error,
// the execution stops, and the error is returned.
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
