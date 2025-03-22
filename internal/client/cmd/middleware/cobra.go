package middleware

import "github.com/spf13/cobra"

type CobraRunE func(cmd *cobra.Command, args []string) error

type MW func(CobraRunE) CobraRunE
