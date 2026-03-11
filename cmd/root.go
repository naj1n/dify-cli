package cmd

import (
	"fmt"
	"os"

	"dify-cli/pkg/client"
	"dify-cli/pkg/config"

	"github.com/spf13/cobra"
)

var (
	flagApp string
	flagKey string
)

var rootCmd = &cobra.Command{
	Use:   "dify",
	Short: "Dify CLI - interact with Dify workflow APIs",
	Long: `Dify CLI is a command-line tool for interacting with Dify managed
or self-hosted instances. It supports executing workflows, uploading
files, viewing logs, and more through the Dify Service API.

A single CLI instance can target multiple workflow apps. The host is
shared, while each app has its own API key.

Get started:
  dify config set-host https://your-dify-instance.com
  dify app add my-workflow app-your-api-key
  dify app default my-workflow

Then run a workflow:
  dify run -i '{"key": "value"}' -u user1
  dify run -a other-app -i '{"key": "value"}'
  dify run -k app-direct-key -i '{"key": "value"}'`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// newClient resolves config + API key and returns a ready-to-use client.
// All API-facing commands should call this.
func newClient() (*client.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.ValidateHost(); err != nil {
		return nil, err
	}
	apiKey, err := cfg.ResolveAPIKey(flagKey, flagApp)
	if err != nil {
		return nil, err
	}
	return client.New(cfg.Host, apiKey), nil
}

func addAppFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flagApp, "app", "a", "", "App name (registered via 'dify app add')")
	cmd.Flags().StringVarP(&flagKey, "key", "k", "", "API key (direct, overrides -a)")
}
