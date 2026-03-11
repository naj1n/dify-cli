package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopUser string

var stopCmd = &cobra.Command{
	Use:   "stop <task_id>",
	Short: "Stop a running workflow task",
	Long:  "Stop a running streaming workflow task by its task ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		data, err := c.StopWorkflow(args[0], stopUser)
		if err != nil {
			return fmt.Errorf("failed to stop workflow: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	addAppFlags(stopCmd)
	stopCmd.Flags().StringVarP(&stopUser, "user", "u", "cli-user", "User identifier")
	rootCmd.AddCommand(stopCmd)
}
