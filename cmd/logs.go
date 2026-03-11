package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	logsKeyword string
	logsStatus  string
	logsPage    int
	logsLimit   int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get workflow execution logs",
	Long: `Retrieve workflow execution logs with optional filtering.

Examples:
  dify logs
  dify logs -a my-app --status succeeded --limit 5
  dify logs -k app-xxx --keyword "error" --page 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		params := map[string]string{
			"keyword": logsKeyword,
			"status":  logsStatus,
		}
		if logsPage > 0 {
			params["page"] = strconv.Itoa(logsPage)
		}
		if logsLimit > 0 {
			params["limit"] = strconv.Itoa(logsLimit)
		}

		data, err := c.GetWorkflowLogs(params)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		var result struct {
			Page    int  `json:"page"`
			Limit   int  `json:"limit"`
			Total   int  `json:"total"`
			HasMore bool `json:"has_more"`
			Data    []struct {
				ID          string `json:"id"`
				WorkflowRun struct {
					ID          string  `json:"id"`
					Version     string  `json:"version"`
					Status      string  `json:"status"`
					Error       *string `json:"error"`
					ElapsedTime float64 `json:"elapsed_time"`
					TotalTokens int     `json:"total_tokens"`
					TotalSteps  int     `json:"total_steps"`
					CreatedAt   int64   `json:"created_at"`
					FinishedAt  int64   `json:"finished_at"`
				} `json:"workflow_run"`
				CreatedFrom string `json:"created_from"`
				CreatedAt   int64  `json:"created_at"`
			} `json:"data"`
		}

		if err := json.Unmarshal(data, &result); err != nil {
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf(
			"Page %d/%d (total: %d, has_more: %v)\n\n",
			result.Page,
			(result.Total+result.Limit-1)/max(result.Limit, 1),
			result.Total,
			result.HasMore,
		)

		for _, entry := range result.Data {
			r := entry.WorkflowRun
			statusIcon := "?"
			switch r.Status {
			case "succeeded":
				statusIcon = "OK"
			case "failed":
				statusIcon = "FAIL"
			case "stopped":
				statusIcon = "STOP"
			case "running":
				statusIcon = "RUN"
			}
			fmt.Printf("[%s] %s  %.2fs  %d tokens  %d steps\n",
				statusIcon, r.ID, r.ElapsedTime, r.TotalTokens, r.TotalSteps)
			if r.Error != nil && *r.Error != "" {
				fmt.Printf("       Error: %s\n", *r.Error)
			}
		}

		return nil
	},
}

func init() {
	addAppFlags(logsCmd)
	logsCmd.Flags().StringVar(&logsKeyword, "keyword", "", "Search keyword")
	logsCmd.Flags().
		StringVar(&logsStatus, "status", "", "Filter by status: succeeded/failed/stopped")
	logsCmd.Flags().IntVar(&logsPage, "page", 1, "Page number")
	logsCmd.Flags().IntVar(&logsLimit, "limit", 20, "Items per page")
	rootCmd.AddCommand(logsCmd)
}
