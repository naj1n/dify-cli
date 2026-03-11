package cmd

import (
	"fmt"

	"dify-cli/pkg/config"

	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage workflow app registrations",
	Long: `Register, remove, and list workflow apps. Each app has a name and an API key.
The name is used with -a flag in other commands to target a specific app.

Examples:
  dify app add my-workflow app-xxxxxxxxxxxx
  dify app add another-app app-yyyyyyyyyyyy
  dify app default my-workflow
  dify app list
  dify app remove old-app`,
}

var appAddCmd = &cobra.Command{
	Use:   "add <name> <api_key>",
	Short: "Register a workflow app",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, key := args[0], args[1]
		var msg string
		if err := config.LockedUpdate(func(cfg *config.Config) error {
			if oldKey, exists := cfg.Apps[name]; exists {
				msg = fmt.Sprintf(
					"Updating app %q (%s → %s)\n",
					name,
					config.MaskKey(oldKey),
					config.MaskKey(key),
				)
			}
			cfg.Apps[name] = key
			if len(cfg.Apps) == 1 {
				cfg.DefaultApp = name
			}
			return nil
		}); err != nil {
			return err
		}
		if msg != "" {
			fmt.Print(msg)
		}

		cfg, _ := config.Load()
		fmt.Printf("App %q added (%s)\n", name, config.MaskKey(key))
		if cfg != nil && cfg.DefaultApp == name {
			fmt.Printf("Set as default app.\n")
		}
		return nil
	},
}

var appRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		var newDefault string
		if err := config.LockedUpdate(func(cfg *config.Config) error {
			if _, ok := cfg.Apps[name]; !ok {
				return fmt.Errorf("app %q not found", name)
			}
			delete(cfg.Apps, name)
			if cfg.DefaultApp == name {
				if remaining := cfg.ListApps(); len(remaining) > 0 {
					cfg.DefaultApp = remaining[0]
					newDefault = remaining[0]
				} else {
					cfg.DefaultApp = ""
				}
			}
			return nil
		}); err != nil {
			return err
		}
		fmt.Printf("App %q removed.\n", name)
		if newDefault != "" {
			fmt.Printf("Default app switched to %q.\n", newDefault)
		}
		return nil
	},
}

var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Apps) == 0 {
			fmt.Println("No apps registered. Run: dify app add <name> <api_key>")
			return nil
		}

		for _, name := range cfg.ListApps() {
			marker := "  "
			if name == cfg.DefaultApp {
				marker = "* "
			}
			fmt.Printf("%s%-20s %s\n", marker, name, config.MaskKey(cfg.Apps[name]))
		}
		return nil
	},
}

var appDefaultCmd = &cobra.Command{
	Use:   "default <name>",
	Short: "Set the default app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.LockedUpdate(func(cfg *config.Config) error {
			if _, ok := cfg.Apps[name]; !ok {
				return fmt.Errorf(
					"app %q not found. Register it first: dify app add %s <api_key>",
					name,
					name,
				)
			}
			cfg.DefaultApp = name
			return nil
		}); err != nil {
			return err
		}
		fmt.Printf("Default app set to %q.\n", name)
		return nil
	},
}

func init() {
	appCmd.AddCommand(appAddCmd, appRemoveCmd, appListCmd, appDefaultCmd)
	rootCmd.AddCommand(appCmd)
}
