package cmd

import (
	"encoding/json"
	"fmt"

	"dify-cli/pkg/config"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show application status and basic information",
	Long: `Retrieve and display basic information about a Dify application.

Examples:
  dify status                  # uses default app
  dify status -a my-app        # specific app
  dify status -k app-xxxxx     # direct API key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		c, err := newClient()
		if err != nil {
			return err
		}

		data, err := c.GetAppInfo()
		if err != nil {
			return fmt.Errorf("failed to get app info: %w", err)
		}

		var info struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Tags        []string `json:"tags"`
			Mode        string   `json:"mode"`
			AuthorName  string   `json:"author_name"`
		}
		if err := json.Unmarshal(data, &info); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		fmt.Println("=== Dify Application Status ===")
		fmt.Printf("Name:        %s\n", info.Name)
		fmt.Printf("Mode:        %s\n", info.Mode)
		fmt.Printf("Author:      %s\n", info.AuthorName)
		if info.Description != "" {
			fmt.Printf("Description: %s\n", info.Description)
		}
		if len(info.Tags) > 0 {
			fmt.Printf("Tags:        %v\n", info.Tags)
		}
		fmt.Printf("Host:        %s\n", cfg.Host)
		fmt.Println("Status:      Connected")

		return nil
	},
}

func init() {
	addAppFlags(statusCmd)
	rootCmd.AddCommand(statusCmd)
}
