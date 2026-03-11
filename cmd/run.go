package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"dify-cli/pkg/client"

	"github.com/spf13/cobra"
)

var (
	runInputs       string
	runUser         string
	runResponseMode string
	runInputFile    string
	runOutput       string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a workflow",
	Long: `Execute a Dify workflow with the specified inputs.

Examples:
  # Blocking mode with default app
  dify run -i '{"query": "hello"}' -u user1

  # Streaming mode targeting a specific app
  dify run -a my-app -i '{"query": "hello"}' -m streaming

  # Direct API key, output to file (good for parallel runs)
  dify run -k app-xxx -i '{"query": "hello"}' -o result.json

  # Load inputs from a JSON file
  dify run -f inputs.json -u user1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		inputs := make(map[string]interface{})

		if runInputFile != "" {
			data, err := os.ReadFile(runInputFile)
			if err != nil {
				return fmt.Errorf("failed to read input file: %w", err)
			}
			if err := json.Unmarshal(data, &inputs); err != nil {
				return fmt.Errorf("failed to parse input file: %w", err)
			}
		}

		if runInputs != "" {
			if err := json.Unmarshal([]byte(runInputs), &inputs); err != nil {
				return fmt.Errorf("failed to parse inputs JSON: %w", err)
			}
		}

		if runResponseMode == "streaming" {
			return runStreaming(c, inputs)
		}
		return runBlocking(c, inputs)
	},
}

func getOutputWriter() (io.Writer, func(), error) {
	if runOutput == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(runOutput)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}
	return f, func() { _ = f.Close() }, nil
}

func runBlocking(c *client.Client, inputs map[string]interface{}) error {
	data, err := c.RunWorkflow(inputs, runUser, "blocking")
	if err != nil {
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	out, closer, err := getOutputWriter()
	if err != nil {
		return err
	}
	defer closer()

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintln(out, string(data))
		return nil
	}

	pretty, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintln(out, string(pretty))
	return nil
}

func runStreaming(c *client.Client, inputs map[string]interface{}) error {
	out, closer, err := getOutputWriter()
	if err != nil {
		return err
	}
	defer closer()

	isFile := runOutput != ""
	stderr := os.Stderr
	if isFile {
		fmt.Fprintf(stderr, "Streaming to %s ...\n", runOutput)
	} else {
		fmt.Fprintln(stderr, "Streaming workflow output...")
		fmt.Fprintln(stderr, "---")
	}

	var streamErr error
	netErr := c.RunWorkflowStream(inputs, runUser, func(event client.SSEEvent) {
		switch event.Event {
		case "workflow_started":
			fmt.Fprintln(stderr, "[workflow started]")

		case "node_started":
			var e struct {
				Data struct {
					Title string `json:"title"`
					Index int    `json:"index"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(event.Data), &e); err == nil {
				fmt.Fprintf(stderr, "[node %d: %s]\n", e.Data.Index, e.Data.Title)
			}

		case "text_chunk":
			var e struct {
				Data struct {
					Text string `json:"text"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(event.Data), &e); err == nil {
				fmt.Fprint(out, e.Data.Text)
			}

		case "node_finished":
			var e struct {
				Data struct {
					Status string `json:"status"`
					Title  string `json:"title"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(event.Data), &e); err == nil &&
				e.Data.Status == "failed" {
				fmt.Fprintf(stderr, "[node failed: %s]\n", e.Data.Title)
			}

		case "workflow_finished":
			var e struct {
				Data struct {
					Status      string  `json:"status"`
					ElapsedTime float64 `json:"elapsed_time"`
					TotalTokens int     `json:"total_tokens"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(event.Data), &e); err == nil {
				fmt.Fprintln(stderr, "\n---")
				fmt.Fprintf(stderr, "[workflow %s | %.2fs | %d tokens]\n",
					e.Data.Status, e.Data.ElapsedTime, e.Data.TotalTokens)
				if e.Data.Status != "succeeded" {
					streamErr = fmt.Errorf("workflow %s", e.Data.Status)
				}
			}

		case "error":
			var e struct {
				Message string `json:"message"`
				Code    string `json:"code"`
			}
			if err := json.Unmarshal([]byte(event.Data), &e); err == nil {
				streamErr = fmt.Errorf("server error: %s (code: %s)", e.Message, e.Code)
				fmt.Fprintf(stderr, "[error: %s]\n", e.Message)
			}

		case "ping":
			// keepalive
		}
	})
	if netErr != nil {
		return netErr
	}
	return streamErr
}

func init() {
	addAppFlags(runCmd)
	runCmd.Flags().StringVarP(&runInputs, "inputs", "i", "", "Input variables as JSON string")
	runCmd.Flags().StringVarP(&runUser, "user", "u", "cli-user", "User identifier")
	runCmd.Flags().
		StringVarP(&runResponseMode, "mode", "m", "blocking", "Response mode: blocking or streaming")
	runCmd.Flags().StringVarP(&runInputFile, "file", "f", "", "Load inputs from a JSON file")
	runCmd.Flags().
		StringVarP(&runOutput, "output", "o", "", "Write output to file (useful for parallel runs)")
	rootCmd.AddCommand(runCmd)
}
