# AGENTS.md

Guidance for AI agents working on this codebase.

## Project Overview

Dify CLI is a command-line tool for interacting with Dify managed or self-hosted instances.
It implements the Dify Workflow App API, supporting workflow execution (blocking & streaming),
file uploads, log retrieval, and more.

One CLI can manage multiple workflow apps under a shared host. Each app is registered
with a name and its own API key.

## Tech Stack

- **Language**: Go
- **CLI Framework**: [cobra](https://github.com/spf13/cobra)
- **No TUI** — pure CLI with stdout/stderr output

## Project Structure

```
dify-cli/
├── main.go                 # Entry point
├── go.mod / go.sum         # Go module dependencies
├── cmd/                    # Command definitions (cobra)
│   ├── root.go             # Root command, help, shared newClient() helper
│   ├── config.go           # config set-host / show
│   ├── app.go              # app add / remove / list / default
│   ├── status.go           # GET /info — show app status
│   ├── run.go              # POST /workflows/run (blocking + streaming, -o output)
│   ├── stop.go             # POST /workflows/tasks/:task_id/stop
│   ├── logs.go             # GET /workflows/logs
│   ├── detail.go           # GET /workflows/run/:workflow_run_id
│   └── upload.go           # POST /files/upload
├── pkg/
│   ├── client/
│   │   └── client.go       # Dify HTTP API client (stateless, concurrent-safe)
│   └── config/
│       └── config.go       # Config persistence (~/.config/dify-cli/config.json)
└── AGENTS.md               # This file
```

## Architecture

### Multi-App Design

- **Host** is global (one Dify instance per config)
- **API keys** are per-app, registered by name
- Commands resolve the API key via priority: `-k` flag > `-a` flag > default app
- Each CLI invocation is fully stateless — safe for parallel execution

### Concurrency

- Config file is read-only during API calls (writes only in `config`/`app` commands)
- No shared state between invocations
- `-o` flag on `run` writes output to a file, enabling parallel workflows:
  ```
  dify run -a app1 -i '{}' -o out1.json &
  dify run -a app2 -i '{}' -o out2.json &
  wait
  ```

## Commands

| Command | Description |
|---|---|
| `dify help` | Show help for any command |
| `dify config set-host <url>` | Set the Dify instance base URL |
| `dify config show` | Display current configuration and all apps |
| `dify app add <name> <key>` | Register a workflow app |
| `dify app remove <name>` | Remove a registered app |
| `dify app list` | List all registered apps |
| `dify app default <name>` | Set the default app |
| `dify status [-a app]` | Show app info (GET /info) |
| `dify run -i '{}' [-a app] [-m streaming] [-o file]` | Execute workflow |
| `dify stop <task_id> [-a app]` | Stop a streaming task |
| `dify detail <run_id> [-a app]` | Get workflow run details |
| `dify logs [-a app] [--status ...]` | List workflow execution logs |
| `dify upload <file> [-a app]` | Upload a file for workflow input |

## Configuration

Config is stored at `~/.config/dify-cli/config.json`:

```json
{
  "host": "https://your-dify-instance.com",
  "apps": {
    "my-workflow": "app-xxxxxxxxxxxx",
    "another-app": "app-yyyyyyyyyyyy"
  },
  "default_app": "my-workflow"
}
```

## API Reference

Base URL: `{host}/v1`
Auth: `Authorization: Bearer {API_KEY}`

Implemented endpoints:
- `POST /workflows/run` — Execute Workflow
- `GET /workflows/run/:workflow_run_id` — Get Workflow Run Detail
- `POST /workflows/tasks/:task_id/stop` — Stop Generate
- `POST /files/upload` — File Upload
- `GET /workflows/logs` — Get Workflow Logs
- `GET /info` — Get Application Basic Information
- `GET /parameters` — Get Application Parameters
- `GET /site` — Get Application WebApp Settings

## Build

```bash
go build -o dify .
```
