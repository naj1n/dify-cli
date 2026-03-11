package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"dify-cli/pkg/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global configuration",
	Long:  "View and modify global settings. The host is shared across all apps.",
}

var setHostCmd = &cobra.Command{
	Use:   "set-host <url>",
	Short: "Set the Dify instance host URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host := strings.TrimRight(args[0], "/")
		u, err := url.Parse(host)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("invalid host URL: must start with http:// or https://")
		}
		if err := config.LockedUpdate(func(cfg *config.Config) error {
			cfg.Host = host
			return nil
		}); err != nil {
			return err
		}
		fmt.Printf("Host set to: %s\n", host)
		return nil
	},
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		host := cfg.Host
		if host == "" {
			host = "(not set)"
		}
		defaultApp := cfg.DefaultApp
		if defaultApp == "" {
			defaultApp = "(not set)"
		}

		fmt.Printf("Host:        %s\n", host)
		fmt.Printf("Default App: %s\n", defaultApp)
		fmt.Printf("Apps:        %d registered\n", len(cfg.Apps))

		if len(cfg.Apps) > 0 {
			fmt.Println()
			for _, name := range cfg.ListApps() {
				marker := "  "
				if name == cfg.DefaultApp {
					marker = "* "
				}
				fmt.Printf("  %s%-16s %s\n", marker, name, config.MaskKey(cfg.Apps[name]))
			}
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(setHostCmd, showConfigCmd)
	rootCmd.AddCommand(configCmd)
}
