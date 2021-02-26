package main

import (
    "github.com/alebcay/quickbrew/pkg/cli"
    "github.com/spf13/cobra"
)

func main() {
    var cmdInstall = &cobra.Command {
        Use: "install [formulae]",
        Short: "Install one or more Homebrew packages",
        Args: cobra.MinimumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            cli.Install(args[0])
        },
    }

    var cmdRemove = &cobra.Command {
        Use: "remove [formulae]",
        Short: "Remove one or more Homebrew packages",
        Args: cobra.MinimumNArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            cli.Remove(args[0])
        },
    }

    var cmdInfo = &cobra.Command {
        Use: "info [formula]",
        Short: "Just a test",
        Args: cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            cli.Info(args[0])
        },
    }

    var cmdPour = &cobra.Command {
        Use: "pour [formula]",
        Short: "Just a test",
        Args: cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            cli.Pour(args[0])
        },
    }

    var rootCmd = &cobra.Command{Use: "quickbrew"}
    rootCmd.AddCommand(cmdInstall, cmdRemove, cmdInfo, cmdPour)
    rootCmd.Execute()
}
