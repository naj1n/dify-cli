package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var uploadUser string

var uploadCmd = &cobra.Command{
	Use:   "upload <file_path>",
	Short: "Upload a file for workflow input",
	Long: `Upload a file to Dify for use as a workflow input variable.
Returns the file ID which can be used in workflow inputs.

Example:
  dify upload document.pdf -u user1
  dify upload image.png -a my-app`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		data, err := c.UploadFile(args[0], uploadUser)
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}

		var result struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Size      int    `json:"size"`
			Extension string `json:"extension"`
			MimeType  string `json:"mime_type"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("File uploaded successfully!\n")
		fmt.Printf("  ID:        %s\n", result.ID)
		fmt.Printf("  Name:      %s\n", result.Name)
		fmt.Printf("  Size:      %d bytes\n", result.Size)
		fmt.Printf("  Extension: %s\n", result.Extension)
		fmt.Printf("  MIME Type: %s\n", result.MimeType)

		return nil
	},
}

func init() {
	addAppFlags(uploadCmd)
	uploadCmd.Flags().StringVarP(&uploadUser, "user", "u", "cli-user", "User identifier")
	rootCmd.AddCommand(uploadCmd)
}
