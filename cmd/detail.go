package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var detailCmd = &cobra.Command{
	Use:   "detail <workflow_run_id>",
	Short: "Get workflow run details",
	Long:  "Retrieve the execution result of a specific workflow run by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		data, err := c.GetWorkflowRunDetail(args[0])
		if err != nil {
			return fmt.Errorf("failed to get workflow detail: %w", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			fmt.Println(string(data))
			return nil
		}
		pretty, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(pretty))
		return nil
	},
}

func init() {
	addAppFlags(detailCmd)
	rootCmd.AddCommand(detailCmd)
}
