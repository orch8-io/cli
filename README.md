# orch8 CLI

Command-line interface for the [Orch8](https://orch8.io) workflow engine. Manage sequences, instances, cron schedules, triggers, and more.

## Install

### Homebrew (macOS / Linux)

```bash
brew tap orch8-io/orch8
brew install orch8-cli
```

### From source

```bash
go install github.com/orch8-io/cli@latest
```

### Binary releases

Download pre-built binaries from [GitHub Releases](https://github.com/orch8-io/cli/releases).

## Configuration

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--url` | `ORCH8_URL` | `http://localhost:8080` | Engine API URL |
| `--tenant` | `ORCH8_TENANT` | `default` | Tenant ID |
| `--api-key` | `ORCH8_API_KEY` | | API key for authentication |
| `--json` | | `false` | Output as JSON |
| `--verbose` | | `false` | Print request details |

```bash
# Configure via environment
export ORCH8_URL=http://localhost:8080
export ORCH8_TENANT=default
export ORCH8_API_KEY=your-api-key
```

## Quick start

```bash
# Check engine health
orch8 health

# Deploy a sequence definition
orch8 deploy sequence.json

# Create a workflow instance
orch8 instance create --sequence <sequence-id>

# Watch instance progress
orch8 instance get <instance-id>
orch8 instance tree <instance-id>
```

## Commands

### deploy

Deploy a sequence definition from a JSON file.

```bash
orch8 deploy <file>
```

### sequence (alias: seq)

Manage sequence definitions.

```bash
orch8 sequence get <id>
orch8 sequence get-by-name <namespace> <name>
orch8 sequence versions <namespace> <name>
orch8 sequence deprecate <id>
```

### instance (aliases: inst, i)

Manage workflow instances.

```bash
# List and inspect
orch8 instance list [--sequence <id>] [--state <state>] [--limit 50]
orch8 instance get <id>
orch8 instance tree <id>
orch8 instance outputs <id>
orch8 instance audit <id>
orch8 instance dlq

# Create
orch8 instance create --sequence <id> [--context '{"key":"value"}'] [--context-file ctx.json] [--idempotency-key <key>]

# Lifecycle
orch8 instance pause <id>
orch8 instance resume <id>
orch8 instance cancel <id>
orch8 instance retry <id>
orch8 instance signal <id> <signal-type> [--payload '{}']
```

### session

Manage stateful sessions that group instances.

```bash
orch8 session create --key <key> [--data '{}']
orch8 session get <id>
orch8 session get-by-key <key>
orch8 session instances <session-id>
orch8 session close <id>
```

### trigger

Manage event triggers.

```bash
orch8 trigger list
orch8 trigger get <slug>
orch8 trigger create --file trigger.json
orch8 trigger fire <slug> [--payload '{}']
orch8 trigger delete <slug>
```

### cron

Manage scheduled workflow execution.

```bash
orch8 cron list
orch8 cron get <id>
orch8 cron create --file schedule.json
orch8 cron enable <id>
orch8 cron disable <id>
orch8 cron delete <id>
```

### pool

Manage resource pools for concurrency control.

```bash
orch8 pool list
orch8 pool get <id>
orch8 pool create --file pool.json
orch8 pool resources <pool-id>
orch8 pool delete <id>
```

### worker

Monitor external worker tasks.

```bash
orch8 worker list
orch8 worker stats
```

### credential

Manage stored credentials.

```bash
orch8 credential list
orch8 credential get <id>
orch8 credential create --file cred.json
orch8 credential delete <id>
```

### plugin

Manage engine plugins.

```bash
orch8 plugin list
orch8 plugin get <name>
orch8 plugin create --file plugin.json
orch8 plugin delete <name>
```

### cluster

Manage engine cluster nodes.

```bash
orch8 cluster list
orch8 cluster drain <node-id>
```

### approval

Manage human-in-the-loop approvals.

```bash
orch8 approval list
```

### circuit-breaker (alias: cb)

Manage circuit breaker states.

```bash
orch8 cb list
orch8 cb get <handler>
orch8 cb reset <handler>
```

### Shell completion

```bash
# Bash
orch8 completion bash > /etc/bash_completion.d/orch8

# Zsh
orch8 completion zsh > "${fpath[1]}/_orch8"

# Fish
orch8 completion fish > ~/.config/fish/completions/orch8.fish
```

## License

BUSL-1.1
